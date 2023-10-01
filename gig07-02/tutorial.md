# WORK IN PROGRESS

# How to connect a Node.js application on Cloud Run to a Cloud SQL for PostgreSQL database

# 1. Overview
The [Cloud SQL Node.js connector](https://github.com/GoogleCloudPlatform/cloud-sql-nodejs-connector#readme) is the easiest way to securely connect your Node.js application to your Cloud SQL database. [Cloud Run](https://cloud.google.com/run) is a fully managed serverless platform that enables you to run stateless containers that are invocable via HTTP requests. This Codelab will demonstrate how to connect a Node.js application on Cloud Run to a Cloud SQL for PostgreSQL database securely with a service account using IAM Authentication.

## What you will learn

In this lab, you will learn how to do the following:

- Create a Cloud SQL instance for PostgreSQL database
- Deploy a Node.js application to Cloud Run
- Connect your application to your database using the Cloud SQL Node.js Connector library

## Prerequisites

This lab assumes familiarity with the Cloud Console and Cloud Shell environments.

# 2. Before you begin

## Cloud Project setup

TODO: pull the contents from the previous tutorial

## Environment Setup

Activate Cloud Shell by clicking on the icon to the right of the search bar.

From Cloud Shell, enable the APIs:

```bash
gcloud services enable compute.googleapis.com sqladmin.googleapis.com \
  run.googleapis.com artifactregistry.googleapis.com \
  cloudbuild.googleapis.com servicenetworking.googleapis.com
```

This command may take a few minutes to complete, but it should eventually produce a successful message similar to this one:

```bash
Operation "operations/acf.p2-327036483151-73d90d00-47ee-447a-b600-a6badf0eceae" finished successfully.
```

# 3. Set up a Service Account

Create and configure a Google Cloud service account to be used by Cloud Run so that it has the correct permissions to connect to Cloud SQL.

1. Run the gcloud iam service-accounts create command as follows to create a new service account:

```bash
gcloud iam service-accounts create quickstart-service-account \
  --display-name="Quickstart Service Account"
```

2. Run the gcloud projects add-iam-policy-binding command as follows to add the Cloud SQL Client role to the Google Cloud service account you just created. In Cloud Shell, the expression ${GOOGLE_CLOUD_PROJECT} will be replaced by the name of your project. You can also do this replacement manually if you feel more comfortable with that.

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member="serviceAccount:quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"
```

3. Run the gcloud projects add-iam-policy-binding command as follows to add the **Cloud SQL Instance User** role to the Google Cloud service account you just created.

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member="serviceAccount:quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --role="roles/cloudsql.instanceUser"

```

4. Run the gcloud projects add-iam-policy-binding command as follows to add the **Log Writer** role to the Google Cloud service account you just created.

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member="serviceAccount:quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --role="roles/logging.logWriter"
```

# 4. Set up Cloud SQL
Run the `gcloud sql instances create` command to create a Cloud SQL instance.

– **-database-version**: The database engine type and version. If left unspecified, the API default is used. See the gcloud database versions documentation to see the current available versions.
– **-cpu**: The number of cores desired in the machine.
– **-memory**: Whole number value indicating how much memory is desired in the machine. A size unit should be provided (for example, 3072MB or 9GB). If no units are specified, GB is assumed.
- **–region**: Regional location of the instance (for example: us-central1, asia-east1, us-east1).
- **–database-flags**: Allows setting flags. In this case, we are turning on cloudsql.iam_authentication to enable Cloud Run to connect to Cloud SQL using the service account we created before.

```bash
gcloud sql instances create quickstart-instance \
  --database-version=POSTGRES_14 \
  --cpu=1 \
  --memory=4GB \
  --region=us-central1 \
  --database-flags=cloudsql.iam_authentication=on

```

This command may take a few minutes to complete.

Run the `gcloud sql databases create` command to create a Cloud SQL database within the `quickstart-instance`.

```bash
gcloud sql databases create quickstart_db \
  --instance=quickstart-instance
```

Create a PostgreSQL database user for the service account you created earlier to access the database.

```bash
gcloud sql users create quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam \
  --instance=quickstart-instance \
  --type=cloud_iam_service_account
```

# 5. Prepare Application

Prepare a Node.js application that responds to HTTP requests.

1. In Cloud Shell create a new directory named `helloworld`, then change into that directory:

```bash
mkdir helloworld
cd helloworld
```

2. Initialize a `package.json` file as a module.
```bash
npm init -y
npm pkg set type="module"
npm pkg set main="index.mjs"
npm pkg set scripts.start="node index.mjs"
```

3. Install the Cloud SQL Node.js connector dependency.

```bash
npm install @google-cloud/cloud-sql-connector
```

4. Install `pg` to interact with the PostgreSQL database.

```bash
npm install pg
```

5. Install express accept incoming http requests.

```bash
npm install express
```

6. Create an `index.mjs` file with the application code. This code is able to:

- Accept HTTP requests
- Connect to the database
- Store the time of the HTTP request in the database
- Return the times of the last five requests

Run the following command in Cloud Shell:

```bash
cat > index.mjs << "EOF"
import express from 'express';
import pg from 'pg';
import {Connector} from '@google-cloud/cloud-sql-connector';

const {Pool} = pg;

const connector = new Connector();
const clientOpts = await connector.getOptions({
    instanceConnectionName: process.env.INSTANCE_CONNECTION_NAME,
    authType: 'IAM'
});

const pool = new Pool({
    ...clientOpts,
    user: process.env.DB_USER,
    database: process.env.DB_NAME
});

const app = express();

app.get('/', async (req, res) => {
  await pool.query('INSERT INTO visits(created_at) VALUES(NOW())');
  const {rows} = await pool.query('SELECT created_at FROM visits ORDER BY created_at DESC LIMIT 5');
  console.table(rows); // prints the last 5 visits
  res.send(rows);
});

const port = parseInt(process.env.PORT) || 8080;
app.listen(port, async () => {
  console.log('process.env: ', process.env);
  await pool.query(`CREATE TABLE IF NOT EXISTS visits (
    id SERIAL NOT NULL,
    created_at timestamp NOT NULL,
    PRIMARY KEY (id)
  );`);
  console.log(`helloworld: listening on port ${port}`);
});

EOF
```

This code creates a basic web server that listens on the port defined by the PORT environment variable. The application is now ready to be deployed.

# 6. Deploy Cloud Run Application

Run the command below to deploy your application to Cloud Run:

- **–region**: Regional location of the instance (for example: us-central1, asia-east1, us-east1).
- **–source**: The source code to be deployed. In this case, . refers to the source code in the current folder helloworld.
- **–set-env-vars**: Sets environment variables used by the application to direct the application to the Cloud SQL database.
- **–service-account**: Ties the Cloud Run deployment to the service account with permissions to connect to the Cloud SQL database created at the beginning of this Codelab.
- **–allow-unauthenticated**: Allows unauthenticated requests so that the application is accessible from the internet.

```bash
gcloud run deploy helloworld \
  --region=us-central1 \
  --source=. \
  --set-env-vars INSTANCE_CONNECTION_NAME="${GOOGLE_CLOUD_PROJECT}:us-central1:quickstart-instance" \
  --set-env-vars DB_NAME="quickstart_db" \
  --set-env-vars DB_USER="quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam" \
  --service-account="quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --allow-unauthenticated
```

If prompted, press `**y**` and `**Enter**` to confirm that you would like to continue:

```bash
Do you want to continue (Y/n)? y
```

After a few minutes, the application should provide a URL for you to visit.

Navigate to the URL to see your application in action. Every time you visit the URL or refresh the page, you will see the five most recent visits returned as JSON.

# 7. Congratulations

You have deployed a Node.js application on Cloud Run taht is able to connect to a PostgreSQL database running on Cloud SQL.

## What we've covered:
- Creating a Cloud SQL for PostgreSQL database
- Deploying a Node.js application to Cloud Run
- Connecting your application to Cloud SQL using the Cloud SQL Node.js Connector

## Clean up
To avoid incurring charges to your Google Cloud account for the resources used in this tutorial, either delete the project that contains the resources, or keep the project and delete the individual resources. If you would like to delete the entire project, you can run:

```bash
gcloud projects delet ${GOOGLE_CLOUD_PROJECT}
```
