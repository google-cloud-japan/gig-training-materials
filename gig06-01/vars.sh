export PROJECT_ID=$(gcloud config get-value project)
export REGION=asia-northeast1

# set the Cloud Run (fully managed) region
$(gcloud config set run/region $REGION)

# set the metrics-writer URL, ignoring if not present
export WRITER_URL=$(gcloud run services describe metrics-writer --format='value(status.url)' 2> /dev/null)
