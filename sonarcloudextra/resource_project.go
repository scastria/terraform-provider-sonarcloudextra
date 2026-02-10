package sonarcloudextra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scastria/terraform-provider-sonarcloudextra/sonarcloudextra/client"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"installation_keys": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"use_existing": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { return d.Id() != "" },
			},
		},
	}
}

func fillProject(c *client.Project, d *schema.ResourceData) {
	c.Organization = d.Get("organization").(string)
	c.Name = d.Get("name").(string)
	c.ProjectKey = fmt.Sprintf("%s_%s", c.Organization, c.Name)
	c.InstallationKeys = d.Get("installation_keys").(string)
	c.UseExisting = d.Get("use_existing").(bool)
}

func fillResourceDataFromProject(c *client.Project, d *schema.ResourceData) {
	d.Set("organization", c.Organization)
	d.Set("name", c.Name)
	d.Set("project_key", c.ProjectKey)
	d.Set("installation_keys", c.InstallationKeys)
	d.Set("use_existing", c.UseExisting)
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	newProject := client.Project{}
	fillProject(&newProject, d)
	var body *bytes.Buffer = nil
	var err error
	if newProject.UseExisting {
		query := url.Values{
			"organization": []string{newProject.Organization},
			"projects":     []string{newProject.ProjectKey},
		}
		body, err = c.HttpRequest(ctx, http.MethodGet, client.ProjectSearchPath, query, nil, &bytes.Buffer{})
		if err != nil {
			re := err.(*client.RequestError)
			if re.StatusCode != http.StatusNotFound {
				return diag.FromErr(err)
			}
			body = nil
		} else {
			searchResp := &client.ProjectSearchResponse{}
			derr := json.NewDecoder(body).Decode(searchResp)
			if derr != nil {
				d.SetId("")
				return diag.FromErr(derr)
			}
			if len(searchResp.Components) == 0 {
				body = nil
			} else {
				tmp := bytes.Buffer{}
				_ = json.NewEncoder(&tmp).Encode(searchResp.Components[0])
				body = &tmp
			}
		}
	}
	if body == nil {
		form := url.Values{
			"installationKeys": []string{newProject.InstallationKeys},
			"organization":     []string{newProject.Organization},
		}
		requestHeaders := http.Header{headers.ContentType: []string{client.FormUrlEncoded}}
		buf := bytes.NewBufferString(form.Encode())
		_, err = c.HttpRequest(ctx, http.MethodPost, client.AlmProvisionProjectsPath, nil, requestHeaders, buf)
		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}
		vquery := url.Values{
			"organization": []string{newProject.Organization},
		}
		var linkedKey string
		for i := 0; i < 10; i++ {
			vbody, verr := c.HttpRequest(ctx, http.MethodGet, client.AlmListRepositoriesPath, vquery, nil, &bytes.Buffer{})
			if verr != nil {
				d.SetId("")
				return diag.FromErr(verr)
			}
			reposResp := &client.AlmListRepositoriesResponse{}
			vderr := json.NewDecoder(vbody).Decode(reposResp)
			if vderr != nil {
				d.SetId("")
				return diag.FromErr(vderr)
			}
			for _, r := range reposResp.Repositories {
				if r.InstallationKey == newProject.InstallationKeys {
					if len(r.LinkedProjects) > 0 {
						linkedKey = r.LinkedProjects[0].Key
					}
					break
				}
			}
			if linkedKey != "" {
				break
			}

			time.Sleep(2 * time.Second)
		}
		if linkedKey != "" {
			newProject.ProjectKey = linkedKey
		}
		fillResourceDataFromProject(&newProject, d)
		d.SetId(newProject.ProjectKey)
		return diags
	}
	retVal := &client.ProjectComponent{}
	err = json.NewDecoder(body).Decode(retVal)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	newProject.ProjectKey = retVal.Key
	fillResourceDataFromProject(&newProject, d)
	d.SetId(retVal.Key)
	return diags
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	projectKey := d.Id()
	org := d.Get("organization").(string)
	if org == "" && projectKey != "" {
		parts := strings.SplitN(projectKey, "_", 2)
		if len(parts) > 0 {
			org = parts[0]
			_ = d.Set("organization", org)
		}
	}
	query := url.Values{
		"organization": []string{org},
		"projects":     []string{projectKey},
	}
	body, err := c.HttpRequest(ctx, http.MethodGet, client.ProjectSearchPath, query, nil, &bytes.Buffer{})
	if err != nil {
		d.SetId("")
		re := err.(*client.RequestError)
		if re.StatusCode == http.StatusNotFound {
			return diags
		}
		return diag.FromErr(err)
	}
	searchResp := &client.ProjectSearchResponse{}
	derr := json.NewDecoder(body).Decode(searchResp)
	if derr != nil {
		d.SetId("")
		return diag.FromErr(derr)
	}
	if len(searchResp.Components) == 0 {
		d.SetId("")
		return diags
	}
	projectKey = searchResp.Components[0].Key
	d.SetId(projectKey)
	if d.Get("installation_keys").(string) == "" || d.Get("name").(string) == "" {
		rq := url.Values{
			"organization": []string{org},
		}
		rbody, rerr := c.HttpRequest(ctx, http.MethodGet, client.AlmListRepositoriesPath, rq, nil, &bytes.Buffer{})
		if rerr != nil {
			return diag.FromErr(rerr)
		}
		reposResp := &client.AlmListRepositoriesResponse{}
		rderr := json.NewDecoder(rbody).Decode(reposResp)
		if rderr != nil {
			return diag.FromErr(rderr)
		}
		for _, r := range reposResp.Repositories {
			for _, lp := range r.LinkedProjects {
				if lp.Key == projectKey {
					_ = d.Set("installation_keys", r.InstallationKey)
					_ = d.Set("name", r.Label)
					_ = d.Set("use_existing", true)
					return diags
				}
			}
		}
	}
	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceProjectRead(ctx, d, m)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	projectKey := d.Id()
	query := url.Values{
		"project": []string{projectKey},
	}
	_, err := c.HttpRequest(ctx, http.MethodPost, client.ProjectsDeletePath, query, nil, &bytes.Buffer{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
