
locals {
  project = "" // Set GCP Project name here
  image   = "squwid/bgcs-site-proxy:v0.1.2"
}

# Create a private static website Google Storage Bucket
resource "google_storage_bucket" "website_bucket" {
  name          = "static-website-files"
  location      = "US-CENTRAL1"
  force_destroy = false

  uniform_bucket_level_access = true
}

resource "google_cloud_run_service" "backend_service" {
  name     = "bytegolf-images-backend"
  location = "us-central1"

  metadata {
    annotations = {
      "run.googleapis.com/ingress" : "all"
    }
  }

  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = "0"
        "autoscaling.knative.dev/maxScale" = "3"
      }
    }

    spec {
      service_account_name  = google_service_account.cloud_run_service_account.email
      container_concurrency = 5
      timeout_seconds       = 30

      containers {
        image = local.image

        resources {
          limits = {
            memory = "256Mi"
            cpu    = "1000m"
          }
        }

        ports {
          container_port = "8000"
        }

        env {
          name  = "BG_BUCKET_NAME"
          value = google_storage_bucket.website_bucket.name
        }

        # File to show when requested file not found. If not set, returns 400
        # status code.
        env {
          name  = "BGCS_NOT_FOUND_FILE"
          value = ""
        }

        # File to show when an endpoint is requested.
        # Ex. /about -> /about/{default_file} 
        # or / -> /{default_file}.
        env {
          name  = "BGCS_DEFAULT_FILE"
          value = "index.html"
        }
      }
    }
  }

  # Split Traffic - https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_service#example-usage---cloud-run-service-traffic-split
  traffic {
    percent         = 100
    latest_revision = true
  }
  autogenerate_revision_name = true

  depends_on = [
    google_service_account.cloud_run_service_account
  ]
}

# OPTIONAL: Allow ALL users to see the files from cloud run.
data "google_iam_policy" "static_site_noauth" {
  binding {
    role    = "roles/run.invoker"
    members = ["allUsers"]
  }
}

resource "google_cloud_run_service_iam_policy" "static_site_noauth_policy" {
  location = google_cloud_run_service.backend_service.location
  project  = google_cloud_run_service.backend_service.project
  service  = google_cloud_run_service.backend_service.name

  policy_data = data.google_iam_policy.static_site_noauth.policy_data
}

resource "google_service_account" "cloud_run_service_account" {
  account_id   = "static-site-service-account"
  display_name = "Static site cloud run service account"
}

# Allow for read access from the Cloud Run to the GCP bucket.
resource "google_storage_bucket_iam_member" "bucket_object_viewer" {
  bucket = google_storage_bucket.website_bucket.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.cloud_run_service_account.email}"
}
