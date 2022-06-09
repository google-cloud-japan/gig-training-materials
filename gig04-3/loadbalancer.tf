terraform {
  required_version = ">=0.14"
  required_providers {
    google = "~> 3.5"
  }
}

provider "google-beta" {
  project = var.project_id
}

variable "project_id" {
  description = "The GCP project ID"
  type  = string
}

variable "region" {
  default = "asia-northeast1"
}

variable "run_service" {
  default = "metrics-writer"
}

module "enable_lb_apis" {
  source  = "terraform-google-modules/project-factory/google//modules/project_services"
  version = "~> 10.2"
  project_id = var.project_id

  disable_services_on_destroy = false
  activate_apis = [
    "compute.googleapis.com",
  ]
}

module "lb-http" {
  source  = "GoogleCloudPlatform/lb-http/google//modules/serverless_negs"
  version = "~> 6.0.1"
  name    = "lb-http"
  project = var.project_id

  ssl                             = false
  https_redirect                  = false

  backends = {
    default = {
      description = "Cloud Run serverless backend"
      groups = [
        {
          group = google_compute_region_network_endpoint_group.serverless_neg.id
        }
      ]
      enable_cdn             = false
      security_policy        = null
      custom_request_headers = null
      custom_response_headers = null

      iap_config = {
        enable               = false
        oauth2_client_id     = ""
        oauth2_client_secret = ""
      }
      log_config = {
        enable      = false
        sample_rate = null
      }
    }
  }
  depends_on = [module.enable_lb_apis]
}

resource "google_compute_region_network_endpoint_group" "serverless_neg" {
  provider              = google-beta
  name                  = "serverless-neg"
  network_endpoint_type = "SERVERLESS"
  region                = var.region

  cloud_run {
    service = var.run_service
  }
  depends_on = [module.enable_lb_apis]
}
