# GIG ハンズオン (Cloud Native)

## Google Cloud プロジェクトの選択

ハンズオンを行う Google Cloud プロジェクトを作成し、 Google Cloud プロジェクトを選択して **Start/開始** をクリックしてください。

**なるべく新しいプロジェクトを作成してください。**

<walkthrough-project-setup>
</walkthrough-project-setup>

## [解説] ハンズオンの内容

### **概要**
In this lab you implement some core cloud nadtive development principles using Cloud Run. The lab is divided into sections. In each section you configure Cloud Run services to demonstrate a particular cloud native principle.

As defined by the Cloud Native Computing Foundation (CNCF): "Cloud native technologies empower organisations to build and run scalable applications in modern, dynamic environments such as public, private, and hybrid clouds. Containers, service meshes, microservices, immutable infrastructure, and declarative APIs exemplify this approach. These techniques enable loosely coupled systems that are resilient, manageable, and observable. Combined with robust automation, they allow engineers to make high-impact changes frequently and predictably with minimal toil."

The following diagram describes the starting state of the lab. The architecture is fully serverless. You deploy containerised web services to Cloud Run that interact with a Cloud Firestore NoSQL database.

![](./image/overview-img.png)

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

```bash
git clone https://github.com/google-cloud-japan/gig-training-materials.git
```

3. Change into the repo directory and checkout the main branch
```bash
cd gig04-3
```
<!-- シェルの中のリージョンを変更する必要あり <- done -->
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

<!-- Source = https://source.cloud.google.com/cnaw-workspace/cloudrun-visualizer/+/master:README.md -->
1. Download the metrics-writer container image to your local Cloud Shell.
```bash
docker pull asia-northeast1-docker.pkg.dev/gig4-3/gig4-3/metrics-writer:latest
```

2. Run the image. You set an environment variable for your project ID, and map a local port to the container port.
```bash
docker run \
  -e GOOGLE_CLOUD_PROJECT=${PROJECT_ID} \
  -p 8080:8080 \
  asia-northeast1-docker.pkg.dev/gig4-3/gig4-3/metrics-writer:latest
```

you see output like below

**Output (do not copy)**
```
> hello-world-metrics@0.0.1 start /usr/src/app
> functions-framework --target=helloMetrics --source ./src/

Serving function...
Function: helloMetrics
Signature type: http
URL: http://localhost:8080/
```

3. Open a new Cloud Shell tab

4. In the new CLoud Shell tab, call the local container.
```bash
curl localhost:8080
```

You see output like below, indicating a successful response.

**Output (do not copy)**
```
Hello from blue
```

5. Go back to the first Cloud Shell tab. You see logging output from the container, indicating that it is starting metrics schedule. You can ignore these logs.

**Output (do not copy)**
```
URL: http://localhost:8080/
initialising instance: 7229f512-6676-4211-90ca-80545c26aeb1
starting metrics schedule...
Metrics: id=7229f512, activeRequests=0,  requestsSinceLast=1
Metrics: id=7229f512, activeRequests=0,  requestsSinceLast=0
```

>Note: If you are getting an error then wait for a while and re-execute the commands from above steps (Step 2 to Step 5).

6. Stop the locally running container using control-c.

### Deploy the initial architecture
1. Deploy the `metrics-writer` app to Cloud Run. You use a prebuild container image from a Google Artifact Registry.
```bash
gcloud run deploy metrics-writer \
  --concurrency 1 \
  --allow-unauthenticated \
  --image asia-northeast1-docker.pkg.dev/gig4-3/gig4-3/metrics-writer:latest
```

You see output like below

**Output (do not copy)**
```
Deploying container to Cloud Run service [metrics-writer] in project [gig4-3] region [asia-northeast1]
OK Deploying new service... Done.
  OK Creating Revision... Revision deployment finished. Checking container health.
  OK Routing traffic...
  OK Setting IAM Policy...
Done.
Service [metrics-writer] revision [metrics-writer-00001-ras] has been deployed and is serving 100 percent of traffic.
Service URL: https://metrics-writer-rmclwajz3a-an.a.run.app
```

2. Set a shell variable with the value of the URL for the metrics-writer service
```bash
export WRITER_URL=$(gcloud run services describe metrics-writer --format='value(status.url)')
```

3. Verify that you can interact with the metrics-writer service. Replace [SERVICE_URL] with the Service URL value from the output of the previous command.
```bash
curl $WRITER_URL
```

You see output like below

**Output (do not copy)**
```
Hello from blue
```

4. Deploy the `visualizer` app to Cloud Run. Again, you use a prebuilt container image from a Google Artifact Registry.
```bash
gcloud run deploy visualizer \
  --allow-unauthenticated \
  --max-instances 5 \
  --image asia-northeast1-docker.pkg.dev/gig4-3/gig4-3/visualizer:latest
```

5. The visualizer service is a web app. On your local machine, open a web browser to the Service URL, copying the URL value from the output of the deploy command.

You see an empty graph, similar to below:

![](./image/visualizer_graph.png)

6. List the Cloud Run services. You see two services, metrics-writer and visualizer.
```bash
gcloud run services list
```

You see output like below

**Output (do not copy)**
```
✔
SERVICE: metrics-writer
REGION: asia-northeast1
URL: https://metrics-writer-rmclwajz3a-an.a.run.app
LAST DEPLOYED BY: admin@hiroyukimomoi.altostrat.com
LAST DEPLOYED AT: 2022-06-09T02:12:43.845980Z

✔
SERVICE: visualizer
REGION: asia-northeast1
URL: https://visualizer-rmclwajz3a-an.a.run.app
LAST DEPLOYED BY: admin@hiroyukimomoi.altostrat.com
LAST DEPLOYED AT: 2022-06-09T04:38:29.682058Z
```

7. Visit the [Cloud Run section](https://console.cloud.google.com/run) of the cloud console and explore the services.

## 2. Scale-out ready
>**Cloud native principle: cloud native apps are stateless, disposable and engineered for fast, automatic scaling.

In this module you generate request traffic against the metrics-writer Cloud Run service to demonstrate autoscaling behavior. You then modify the service's configuration to see the impact on scaling behavior.

![](./image/scale-out_img.png)

### Cloud Runc container instance autoscaling

In Cloud Run, each active [revision](https://cloud.google.com/run/docs/resource-model#revisions) is automatically scaled to the number of container instances needed to handle incoming requests. Refer to the [instance autoscaling](https://cloud.google.com/run/docs/about-instance-autoscaling) docs for more details.

The number of instances created is impacted by:
- The CPU utilization of existing instances (Targeting to keep serving instances to a 60% CPU utilization)
- The [concurrency setting](https://cloud.google.com/run/docs/about-concurrency)
- The [maximum number of container instances setting](https://cloud.google.com/run/docs/configuring/max-instances)
- The [minumum number of container instances setting](https://cloud.google.com/run/docs/configuring/min-instances)

### Generate request traffic

1. Open Cloud Shell. If your previous shell was inactive for some time, you may need to reconnect. If so, after reconnecting, change into the repo directory and set the environment variables again.
```bash
cd ~/gig-training-materials/gig04-3 && source vars.sh && export WRITER_URL=$(gcloud run services describe metrics-writer --format='value(status.url)')
```

2. List the Cloud Run services.
```bash
gcloud run services list
```

3. If you do not have one open already, one a web browser page to the url of the visualizer service.

4. Use the [hey](https://github.com/rakyll/hey) command-line utility to generate request traffic against the service for 30 seconds, using 30 workers. The `hey` utility is already installed in Cloud Shell.
```bash
hey -z 30s -c 30 $WRITER_URL
```

5. Switch to the browser page that displays the visualizer web app. You see a graph plotted on the page. Cloud Run has rapidly scaled the number of active instances to serve the traffic volume.

![](./image/visualizer_graph_2.png)

6. Watch the graph until the end of 30 seconds. Cloud Run rapidly scales down to zero instances. Make a mental note of the peak number of active instances.

7. Return to cloud shell. The `hey` utility outputs a summary of the load test. Look at the summary metrics and response time histogram.

![](./image/hey_summary.png)

8. Visit the [Cloud Run section](https://console.cloud.google.com/run) of the cloud console. Click into the `metrics-writer` service, and then select the 'Metrics' tab.

<image />

You see that Cloud Run provides some useful [monitoring metrics] out-of-the-box, such as request count, request latencies, container instance count, and more.

9. Change the time period to '1 hour' and look at the 'container instance count' graph. The peak 'active' instances value should match approximately the value you saw in the visualizer graph. You need to wait approximately 3 minutes for the graphs to update.

>Note: the Metrics tab in the Cloud Console provides the most accurate information about your Cloud Run service. This information comes from Cloud Monitoring. However, the metrics in the console take approximately 3 minutes to update. In this lab, you use the visualizer graph to show real-time scaling. The visualizer is for demo purposes only.

### Update service concurrency

Cloud Run provides a [concurrency]() setting that specifies the maximum number of requests that can be processed simultaneously by a given container instance.

If your code cannot process parallel requests, set `concurrency=1`. Each container instance will handle only 1 request at a time, as in the diagram on the left.

If your container can handle multiple requests simultaneously, set a higher concurrency. The specified concurrency value is a _maximum_ and Cloud Run might not spend as many requests to a given container instance if the CPU of the instance is already highly utilized. In the diagram on the right, the service is configured to handle a maximum of 80 simultaneous requests(the default). Cloud Run therefore sends all 3 requests to a single container instance.

![](./image/concurrency_image.png)

You deployed the metrics-writer service with an initial setting of `concurrency=1`. This means that each container instance will process only a single request at a time. You used this value to demonstrate Cloud Run's fast autoscaling. However, a simple service like this can probably handle a much higher concurrency. Here, you increase the concurrency setting and investigate the impact on scaling behavior.

1. Update the metrics-writer service's concurrency setting. This creates a new revision for the service. All requests are routed to this new revision once it is ready.
```bash
gcloud run services update metrics-writer \
  --concurrency 5
```

2. Rerun the command to generate request load
```bash
hey -z 30s -c 30 $WRITER_URL
```

3. Switch back to the browser page that displays the visualizer web app. You see another graph plotted on the page.

![](./image/visualizer_graph_3.png)

4. Inspect the `hey` output summary.

![](./image/hey_summary_2.png)

### Update service max-instances configuration

Here, you use the [maximum container instances]() setting to limit the scaling of your service in response to incoming requests. Use this setting as a way to control your costs or to limit the number of connections to a backing service, such as to a database.

1. Update the metrics-writer service's max-instances setting
```bash
gcloud run services update metrics-writer \
  --max-instances 5
```

2. Rerun the command to generate request load
```bash
hey -z 30s -c 30 $WRITER_URL
```

3. Switch back to the browser page that displays the visualizer web app. You see another graph plotted on the page.

4. Inspect the `hey` output summary. How does it compare to the previous output?
