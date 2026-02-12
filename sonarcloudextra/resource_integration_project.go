package sonarcloudextra

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-http-utils/headers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scastria/terraform-provider-sonarcloudextra/sonarcloudextra/client"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIntegrationProjectCreate,
		ReadContext:   resourceIntegrationProjectRead,
		DeleteContext: resourceIntegrationProjectDelete,
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
				Required: true,
				ForceNew: true,
			},
			"bitbucket_repo_uuid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"use_existing": {
				Type:             schema.TypeBool,
				Optional:         true,
				ForceNew:         true,
				Default:          false,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { return d.Id() != "" },
			},
		},
	}
}

func fillIntegrationProject(c *client.IntegrationProject, d *schema.ResourceData) {
	c.Organization = d.Get("organization").(string)
	c.Name = d.Get("name").(string)
	c.BitbucketRepoUuid = d.Get("bitbucket_repo_uuid").(string)
	c.UseExisting = d.Get("use_existing").(bool)
}

func fillIntegrationResourceDataFromProjectComponent(c *client.IntegrationProjectComponent, d *schema.ResourceData) {
	d.Set("organization", c.Organization)
	d.Set("name", c.Name)
	d.Set("use_existing", c.UseExisting)
}

func fillIntegrationResourceDataFromAlmRepository(c *client.AlmRepository, d *schema.ResourceData) {
	d.Set("bitbucket_repo_uuid", c.InstallationKey)
}

func resourceIntegrationProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	newIntegrationProject := client.IntegrationProject{}
	fillIntegrationProject(&newIntegrationProject, d)
	foundExisting := false
	if newIntegrationProject.UseExisting {
		query := url.Values{
			"organization": []string{newIntegrationProject.Organization},
			"projects":     []string{client.IntegrationProjectEncodeSonarId(newIntegrationProject.Organization, newIntegrationProject.Name)},
		}
		body, err := c.HttpRequest(ctx, http.MethodGet, client.ProjectSearchPath, query, nil, &bytes.Buffer{})
		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}
		searchResp := &client.IntegrationProjectSearchResponse{}
		err = json.NewDecoder(body).Decode(searchResp)
		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}
		if len(searchResp.Components) == 1 {
			foundExisting = true
		}
	}
	if !foundExisting {
		form := url.Values{
			"organization":           []string{newIntegrationProject.Organization},
			"installationKeys":       []string{newIntegrationProject.BitbucketRepoUuid},
			"newCodeDefinitionType":  []string{"previous_version"},
			"newCodeDefinitionValue": []string{"previous_version"},
		}
		requestHeaders := http.Header{
			headers.ContentType: []string{client.FormUrlEncoded},
		}
		buf := bytes.NewBufferString(form.Encode())
		_, err := c.HttpRequest(ctx, http.MethodPost, client.AlmProvisionProjectsPath, nil, requestHeaders, buf)
		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}
	}
	d.SetId(newIntegrationProject.IntegrationProjectEncodeId())
	return diags
}

func resourceIntegrationProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	organization, name := client.IntegrationProjectDecodeId(d.Id())
	projectKey := client.IntegrationProjectEncodeSonarId(organization, name)
	query := url.Values{
		"organization": []string{organization},
		"projects":     []string{projectKey},
	}
	body, err := c.HttpRequest(ctx, http.MethodGet, client.ProjectSearchPath, query, nil, &bytes.Buffer{})
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	searchResp := &client.IntegrationProjectSearchResponse{}
	err = json.NewDecoder(body).Decode(searchResp)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	if len(searchResp.Components) == 0 {
		d.SetId("")
		return diags
	}
	searchProject := searchResp.Components[0]
	query = url.Values{
		"organization": []string{organization},
	}
	body, err = c.HttpRequest(ctx, http.MethodGet, client.AlmListRepositoriesPath, query, nil, &bytes.Buffer{})
	if err != nil {
		return diag.FromErr(err)
	}
	reposResp := &client.IntegrationAlmListRepositoriesResponse{}
	err = json.NewDecoder(body).Decode(reposResp)
	if err != nil {
		return diag.FromErr(err)
	}
	var almRepository *client.AlmRepository = nil
	for _, r := range reposResp.Repositories {
		for _, lp := range r.LinkedProjects {
			if lp.Key == projectKey {
				almRepository = &r
				break
			}
		}
		if almRepository != nil {
			break
		}
	}
	if almRepository == nil {
		d.SetId("")
		return diags
	}
	fillIntegrationResourceDataFromProjectComponent(&searchProject, d)
	fillIntegrationResourceDataFromAlmRepository(almRepository, d)
	return diags
}

func resourceIntegrationProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	organization, name := client.IntegrationProjectDecodeId(d.Id())
	query := url.Values{
		"project": []string{client.IntegrationProjectEncodeSonarId(organization, name)},
	}
	_, err := c.HttpRequest(ctx, http.MethodPost, client.ProjectsDeletePath, query, nil, &bytes.Buffer{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
