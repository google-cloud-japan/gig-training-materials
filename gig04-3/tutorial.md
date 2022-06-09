# GIG ハンズオン (Cloud Native)

## Google Cloud プロジェクトの選択

ハンズオンを行う Google Cloud プロジェクトを作成し、 Google Cloud プロジェクトを選択して **Start/開始** をクリックしてください。

**なるべく新しいプロジェクトを作成してください。**

<walkthrough-project-setup>
</walkthrough-project-setup>

## [解説] ハンズオンの内容

### **概要**
In this lab you implement some core cloud native development principles using Cloud Run. The lab is divided into sections. In each section you configure Cloud Run services to demonstrate a particular cloud native principle.

As defined by the Cloud Native Computing Foundation (CNCF): "Cloud native technologies empower organisations to build and run scalable applications in modern, dynamic environments such as public, private, and hybrid clouds. Containers, service meshes, microservices, immutable infrastructure, and declarative APIs exemplify this approach. These techniques enable loosely coupled systems that are resilient, manageable, and observable. Combined with robust automation, they allow engineers to make high-impact changes frequently and predictably with minimal toil."

The following diagram describes the starting state of the lab. The architecture is fully serverless. You deploy containerised web services to Cloud Run that interact with a Cloud Firestore NoSQL database.


<overview image />

The architecture consists of two Cloud Run services:

Metrics writer

- Simple 'hello world' style service that writes metrics to a Cloud Firestore database.
- Each metrics-writer instance writes a heartbeat record to the Cloud Firestore database every 1 second.
> The heartbeat record indicates whether the instance is active (processing a request), how many requests it received in the last second, and other metadata
Visualizer web app

- Web app hosted in Cloud Run that reads the metrics persisted by the metrics-writer instances, and displays in a nice graph.

### **目的**
In this lab, you perform the following tasks:

- Deploy containerised services to Cloud Run
- Generate load against Cloud Run to demonstrate scaling behaviour
- Configure a load balancer and traffic splitting rules to manipulate network traffic
- Configure IAM and security rules to limit access to Cloud Run services.

## 1. Containers are universal
> **Cloud native principle**: Containers are the standardised, immutable unit of cloud native software.

In this task you configure your environment and deploy the initial architecture.

- You deploy Cloud Run services using prebuilt container images.
- It doesn't matter what programming language, web frameworks or dependencies the images use.
- The images are packaged in a standardised, universal format.
- The images can be deployed to different container execution environments, without modification.

### Setup your environment
1. Open `Cloud Shell`
2. Clone the Cloud Source Repositories git repository that contains some helper scripts for this lab. If you are requested to authorise gcloud, do so.

<!-- ダウンロードするリポジトリはあとで再検討が必要 -->
```bash
gcloud source repos clone cnaw-participant --project=cnaw-workspace
```

3. Change into the repo directory and checkout the main branch
```bash
cd cnaw-participant
```
<!-- シェルの中のリージョンを変更する必要あり -->
4. Run the helper script to set shell variables for your project ID and default region.
```bash
source vars.sh
```

5. Configure `gcloud` to use Cloud Run manged platform by default
```bash
gcloud config set run/platform managed
```

6. Enable the required APIs
```bash
gcloud services enable run.googleapis.com \
  firestore.googleapis.com \
  appengine.googleapis.com \
  compute.googleapis.com
```

7. Initialize AppEngine. You do not use AppEngine in this lab, but you need to initialize AppEngine before you create a Firestore database in the next step
```bash
gclou dapp create --region $REGION
```

8. Create a FIrestore database.
```bash
gcloud firestore databases create --region $REGION
```

### Run the metrics-writer container locally
Here, you run a metrics-writer container locally. You fetch the container image from a public Google Container Registry. The container image is executable and fully self-contained. You don't need to install any dependencies or runtime environments, as everything is packaged into the image.

<!-- ここでこけるので、source をダウンロードしてどこかに docker image をホストする、かビルドさせる、か。 -->
<!-- Source = https://source.cloud.google.com/cnaw-workspace/cloudrun-visualizer/+/master:README.md -->
1. Download the metrics-writer container image to your local Cloud Shell.

