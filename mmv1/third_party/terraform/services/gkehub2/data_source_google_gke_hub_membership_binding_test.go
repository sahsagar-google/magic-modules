package gkehub2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
	"github.com/hashicorp/terraform-provider-google/google/envvar"
)

func TestAccDataSourceGoogleGKEHub2MembershipBinding_basic(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"org_id":          envvar.GetTestOrgFromEnv(t),
		"billing_account": envvar.GetTestBillingAccountFromEnv(t),
		"random_suffix":   acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckGKEHub2MembershipBindingDestroyProducer(t),
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGoogleGKEHub2MembershipBinding_basic(context),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckDataSourceStateMatchesResourceState("data.google_gke_hub_membership_binding.example", "google_gke_hub_membership_binding.example"),
				),
			},
		},
	})
}

func testAccDataSourceGoogleGKEHub2MembershipBinding_basic(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_project" "project" {
  name            = "tf-test-gkehub%{random_suffix}"
  project_id      = "tf-test-gkehub%{random_suffix}"
  org_id          = "%{org_id}"
  billing_account = "%{billing_account}"
}

resource "google_project_service" "compute" {
  project = google_project.project.project_id
  service = "compute.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "anthos" {
  project = google_project.project.project_id
  service = "anthos.googleapis.com"
}

resource "google_project_service" "gkehub" {
  project = google_project.project.project_id
  service = "gkehub.googleapis.com"
  disable_on_destroy = false
}

resource "time_sleep" "wait_for_gkehub_enablement" {
  create_duration = "150s"
  depends_on = [
  google_project_service.gkehub, 
  google_project_service.compute,
  google_project_service.anthos
  ]
}

resource "google_gke_hub_membership" "example" {
  project = google_project.project.project_id
  membership_id = "tf-test-membership%{random_suffix}"
  depends_on = [time_sleep.wait_for_gkehub_enablement]
}

resource "google_gke_hub_scope" "example" {
  project = google_project.project.project_id
  scope_id = "tf-test-scope%{random_suffix}"
  depends_on = [time_sleep.wait_for_gkehub_enablement]
}

resource "google_gke_hub_membership_binding" "example" {
  project = google_project.project.project_id
  membership_binding_id = "tf-test-membership-binding%{random_suffix}"
  scope = google_gke_hub_scope.example.name
  membership_id = "tf-test-membership%{random_suffix}"
  location = "global"
  labels = {
      keyb = "valueb"
      keya = "valuea"
      keyc = "valuec" 
  }
  depends_on = [
    google_gke_hub_membership.example,
    google_gke_hub_scope.example
  ]
}

data "google_gke_hub_membership_binding" "example" {
  location = google_gke_hub_membership_binding.example.location
  project  = google_gke_hub_membership_binding.example.project
  membership_id = google_gke_hub_membership_binding.example.membership_id
  membership_binding_id = google_gke_hub_membership_binding.example.membership_binding_id
}
`, context)
}
