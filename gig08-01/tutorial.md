# GIG ハンズオン (Cloud Native)

## 環境準備

### Google Cloud プロジェクトの選択

ハンズオンを行う Google Cloud プロジェクトをまだ作成されていない場合は、[こちらのリンク](https://console.cloud.google.com/projectcreate) から新しいプロジェクトを作成してください。

**なるべく新しいプロジェクトが望ましいです。**

### 必要なロール

ハンズオンを進めるためには以下 **1** or **2** の何れかの IAM ロールが必要です。

1. [オーナー](https://cloud.google.com/iam/docs/understanding-roles#basic)
2. [編集者](https://cloud.google.com/iam/docs/understanding-roles#basic)、[Project IAM 管理者](https://cloud.google.com/iam/docs/understanding-roles#resourcemanager.projectIamAdmin)、[Cloud Datastore オーナー](https://cloud.google.com/iam/docs/understanding-roles#datastore.owner)、[Cloud Run 管理者](https://cloud.google.com/iam/docs/understanding-roles#run.admin)

それでは最初に、ハンズオンを進めるための環境準備を行います。

#### GCP のプロジェクト ID を環境変数に設定

環境変数 `PROJECT_ID` に GCP プロジェクト ID を設定します。[GOOGLE_CLOUD_PROJECT_ID] 部分にご使用になられる Google Cloud プロジェクトの ID を入力してください。
例: `export PROJECT_ID=gig7-1`

```bash
export PROJECT_ID=[GOOGLE_CLOUD_PROJECT_ID]
```

#### CLI（gcloud コマンド）から利用する GCP のデフォルトプロジェクトを設定

操作対象のプロジェクトを設定します。

```bash
gcloud config set project $PROJECT_ID
```

デフォルトのリージョンを設定します。

```bash
gcloud config set compute/region asia-northeast1
```

以下のコマンドで、現在の設定を確認できます。
```bash
gcloud config list
```

### ProTips
gcloud コマンドには、config 設定をまとめて切り替える方法があります。
アカウントやプロジェクト、デフォルトのリージョン、ゾーンの切り替えがまとめて切り替えられるので、おすすめの機能です。
```bash
gcloud config configurations list
```

## **参考: Cloud Shell の接続が途切れてしまったときは?**

一定時間非アクティブ状態になる、またはブラウザが固まってしまったなどで `Cloud Shell` が切れてしまう、またはブラウザのリロードが必要になる場合があります。その場合は以下の対応を行い、チュートリアルを再開してください。

### **1. チュートリアル資材があるディレクトリに移動する**

```bash
cd ~/cloudshell_open/gig-training-materials/gig08-01/
```

### **2. チュートリアルを開く**

```bash
teachme tutorial.md
```

### **3. gcloud のデフォルト設定**

```bash
source vars.sh
```

途中まで進めていたチュートリアルのページまで `Next` ボタンを押し、進めてください。

## [解説] ハンズオンの内容

### **概要**
このラボでは、いくつかの重要なクラウドネイティブ開発原則に基づいて Cloud Run を実装します。 ラボは各セクションに分かれています。 各セクションでは、特定のクラウドネイティブ原則を示すように Cloud Run サービスを構成します。

Cloud Native Computing Foundation（CNCF）の定義によると、「クラウドネイティブテクノロジーにより、組織は、パブリック、プライベート、ハイブリッドクラウドなどの最新の動的環境でスケーラブルなアプリケーションを構築および実行できます。コンテナ、サービスメッシュ、マイクロサービス、イミュータブルインフラストラクチャ、および 宣言型 API は、このアプローチの例です。これらの手法により、復元力、管理性、監視性を備えた疎結合システムが実現可能になります。堅牢な自動化と組み合わせることで、エンジニアは最小限の労力で頻繁かつ予測どおりに影響の大きい変更を行うことができます。」

次の図は、ラボの開始状態を示しています。 アーキテクチャは完全にサーバーレスです。 Cloud Firestore NoSQL データベースと相互作用するコンテナ化された Web サービスを Cloud Run にデプロイします。

![](image/overview-img.png?raw=true)

このアーキテクチャは、2つのCloudRunサービスで構成されています。

Metrics writer

- メトリックを Cloud Firestore データベースに書き込むシンプルな「helloworld」スタイルのサービス。
- 各メトリックライターインスタンスは、1秒ごとにハートビートレコードを Cloud Firestore データベースに書き込みます。

> ハートビートレコードは、インスタンスがアクティブであるかどうか（要求を処理しているかどうか）、最後の1秒間に受信した要求の数、およびその他のメタデータを示します

Visualizer web app

- Cloud Run でホストされ、メトリックライターインスタンスによって永続化されたメトリックを読み取り、いい感じのグラフを表示するウェブアプリ。

### **目的**
このラボでは、次のタスクを実行します。

- コンテナ化されたサービスを Cloud Run にデプロイします。
- スケーリング動作を示すために Cloud Run に対して負荷を生成します。
- ネットワークトラフィックを操作するためのロードバランサーとトラフィック分割ルールを構成します。

## 1. Containers はユニバーサル
> _**クラウドネイティブの原則**: コンテナは、クラウドネイティブソフトウェアにおける、標準化されたイミュータブルなユニットです。_

このタスクでは、環境を設定し、最初のアーキテクチャをデプロイします。

- ビルド済みのコンテナイメージを使用して Cloud Run サービスを展開します。
- イメージが使用するプログラミング言語、Webフレームワーク、または依存関係は関係ありません。
- イメージは、標準化されたユニバーサルなフォーマットでパッケージ化されています。
- イメージは、変更することなく、さまざまなコンテナ実行環境に展開できます。

### 環境のセットアップ
1. `Cloud Shell` を開きます。

>Note: README の青い`OPEN IN GOOGLE CLOUD SHELL` ボタンから開始された場合は、すでにリポジトリはクローンされていますので、4 にスキップしてください。

2. このラボのスクリプトを含む git リポジトリをクローンします。 gcloud の承認を求められた場合は、承認してください。
```bash
git clone https://github.com/google-cloud-japan/gig-training-materials.git
```

3. リポジトリディレクトリに移動します。
```bash
cd ~/gig-training-materials/gig08-01
```

4. 必要な API を有効にします。
```bash
gcloud services enable run.googleapis.com \
  firestore.googleapis.com \
  compute.googleapis.com
```

5. スクリプトを実行して、プロジェクト ID とデフォルトリージョンのシェル変数を設定します。
```bash
source vars.sh
```

6. デフォルトで Cloud Run のマネージド環境を利用するよう、 `gcloud` コマンドで設定します。
```bash
gcloud config set run/platform managed
```

7. GCP のプロジェクト番号を環境変数に設定します。

```bash
export PROJECT_NUM=$(gcloud projects describe $PROJECT_ID --format json | jq -r '.projectNumber')
```

8. Cloud Run が使用するサービスアカウントに、プロジェクトの編集者 IAM ロールを設定します。

```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --role roles/editor \
  --member "serviceAccount:$PROJECT_NUM-compute@developer.gserviceaccount.com"
```

9. Firestore データベースを作成します。
```bash
gcloud firestore databases create --location $REGION
```

### metrics-writer コンテナをローカルで実行する
ここでは、metrics-writer コンテナをローカルで実行します。公開されている Google Artifact Registry からコンテナイメージを取得します。コンテナイメージは実行可能であり、完全に自己完結型です。すべてがイメージにパッケージ化されているため、依存関係やランタイム環境をインストールする必要はありません。

<!-- Source = https://source.cloud.google.com/cnaw-workspace/cloudrun-visualizer/+/master:README.md -->
1. metrics-writer コンテナイメージをローカルの Cloud Shell インスタンスにダウンロードします。
```bash
docker pull asia-northeast1-docker.pkg.dev/gig6-1/gig6-1/metrics-writer:latest
```

2. イメージを実行します。プロジェクト ID に環境変数を設定し、ローカルポートをコンテナポートにマップします。
```bash
docker run \
  -e GOOGLE_CLOUD_PROJECT=${PROJECT_ID} \
  -p 8080:8080 \
  asia-northeast1-docker.pkg.dev/gig6-1/gig6-1/metrics-writer:latest
```

以下のような出力が表示されます。

**Output**
```terminal
> hello-world-metrics@0.0.1 start /usr/src/app
> functions-framework --target=helloMetrics --source ./src/

Serving function...
Function: helloMetrics
Signature type: http
URL: http://localhost:8080/
```

3. 新しい Cloud Shell タブを開きます。

4. 新しい Cloud Shell タブで、ローカルコンテナを呼び出します。
```bash
curl localhost:8080
```

以下のような出力が表示され、正常な応答を示しています。

**Output**
```terminal
Hello from blue
```

5. 最初の Cloud Shell タブに戻ります。メトリックスケジュールが開始していることが示されています。これらのログは無視してかまいません。

**Output**
```terminal
URL: http://localhost:8080/
initialising instance: 7229f512-6676-4211-90ca-80545c26aeb1
starting metrics schedule...
Metrics: id=7229f512, activeRequests=0,  requestsSinceLast=1
Metrics: id=7229f512, activeRequests=0,  requestsSinceLast=0
```

>Note: エラーが発生した場合は、しばらく待ってから、上記の手順（手順2から手順5）のコマンドを再実行してください。

6. control-c でローカル実行されているコンテナを停止します。

### 初期アーキテクチャのデプロイ

1. `metrics-writer` アプリを Cloud Run にデプロイします。Google Artifact Registry からのビルド済みコンテナイメージを使用します。
```bash
gcloud run deploy metrics-writer \
  --concurrency 1 \
  --allow-unauthenticated \
  --image asia-northeast1-docker.pkg.dev/gig6-1/gig6-1/metrics-writer:latest
```
以下のような出力が表示されます。

**Output**
```terminal
Deploying container to Cloud Run service [metrics-writer] in project [gig7-1] region [asia-northeast1]
OK Deploying new service... Done.
  OK Creating Revision... Revision deployment finished. Checking container health.
  OK Routing traffic...
  OK Setting IAM Policy...
Done.
Service [metrics-writer] revision [metrics-writer-00001-ras] has been deployed and is serving 100 percent of traffic.
Service URL: https://metrics-writer-rmclwajz3a-an.a.run.app
```

2. metrics-writer サービスの URL の値を使用してシェル変数を設定します。
```bash
export WRITER_URL=$(gcloud run services describe metrics-writer --format='value(status.url)')
```

3. metrics-writer サービスに接続できることを確認します。
```bash
curl $WRITER_URL
```

以下のような出力が表示されます。

**Output**
```terminal
Hello from blue
```

4. `visualizer` アプリを Cloud Run にデプロイします。ここでも、Google Artifact Registry から事前に作成されたコンテナイメージを使用します。
```bash
gcloud run deploy visualizer \
  --allow-unauthenticated \
  --max-instances 5 \
  --image asia-northeast1-docker.pkg.dev/gig6-1/gig6-1/visualizer:latest
```

5. visualizer サービスは Web アプリです。ローカルマシンで、Web ブラウザを開いてサービス URL にアクセスし、deploy コマンドの出力から URL 値をコピーします。

以下のような空のグラフが表示されます:

![](image/visualizer_graph.png?raw=true)

6. Cloud Run サービスを一覧表示します。metrics-writer と visualizer の2つのサービスが表示されます。
```bash
gcloud run services list
```

以下のような出力が表示されます。

**Output**
```terminal
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

7. [Cloud Run セクション](https://console.cloud.google.com/run) にアクセスして、サービスの内容を確認します。

## 2. スケールアウト対応

>_**クラウドネイティブの原則**: クラウドネイティブアプリはステートレスでディスポーザブルであり、高速の自動スケーリング用に設計されています。_

このモジュールでは、metrics-writer の Cloud Run サービスに対してトラフィックを生成して、自動スケーリングの動作を確認します。次に、サービスの構成を変更して、スケーリング動作への影響を確認します。

![](image/scale-out_img.png?raw=true)

### Cloud Run コンテナインスタンスの自動スケーリング

Cloud Runでは、アクティブな各 [リビジョン](https://cloud.google.com/run/docs/resource-model#revisions) は、着信要求を処理するために必要なコンテナインスタンスの数に自動的にスケーリングされます。詳細については、 [インスタンスの自動スケーリング](https://cloud.google.com/run/docs/about-instance-autoscaling) のドキュメントを参照してください。

作成されるインスタンスの数は、次の影響を受けます。
- 既存のインスタンスのCPU使用率（インスタンスを 60％ の CPU 使用率で提供し続けることを目標としています）
- [インスタンスあたりの同時リクエストの最大数（サービス）](https://cloud.google.com/run/docs/about-concurrency)
- [コンテナ インスタンス（サービス）の最大数](https://cloud.google.com/run/docs/configuring/max-instances)
- [最小インスタンス数（サービス）](https://cloud.google.com/run/docs/configuring/min-instances)

### リクエスト トラフィックを生成する

1. Cloud Shell を開きます。以前のシェルがしばらく非アクティブだった場合は、再接続が必要になる場合があります。その場合は、再接続後、リポディレクトリに移動し、環境変数を再設定します。

```bash
cd ~/cloudshell_open/gig-training-materials/gig08-01/ && source vars.sh
```

2. Cloud Run サービスを一覧表示します。
```bash
gcloud run services list
```

3. まだ開いていない場合は、visualizer サービスの URL への Web ブラウザーページを1つ開きます。

4. [hey](https://github.com/rakyll/hey) コマンドラインユーティリティを使用して、サービスに 30 ワーカーで 30 秒間リクエストトラフィックを生成します。 `hey` ユーティリティはすでに Cloud Shell にインストールされています。
```bash
hey -z 30s -c 30 $WRITER_URL
```
オプションは、「30秒間(30s)に同時並列数 30 のリクエストを実行する」ということを意味しています。

5. visualizer Web アプリを表示するブラウザーページに切り替えます。ページにグラフがプロットされています。 Cloud Run は、トラフィック量を処理するためにアクティブなインスタンスの数を急速に拡大しました。

![](image/visualizer_graph_2.png?raw=true)

6. 30 秒が経過するまでグラフを監視します。 Cloud Run は、インスタンスがゼロになるまで急速にスケールダウンします。アクティブなインスタンスのピーク数を覚えておいてください。

7. cloud shell に戻ります。 `hey` ユーティリティは、負荷テストの要約を出力します。要約メトリックと応答時間のヒストグラムを見てください。

![](image/hey_summary.png?raw=true)

8. クラウドコンソールの [Cloud Run セクション](https://console.cloud.google.com/run) にアクセスします。`metrics-writer` サービスをクリックし、`指標` タブを選択します。

![](image/cloudrun_metrics_image.png?raw=true)

Cloud Run は、リクエスト数、リクエストレイテンシ、コンテナインスタンス数など、すぐに使用できる便利な[モニタリング指標]を提供していることがわかります。

9. 期間を「1時間」に変更し、「コンテナ インスタンス 数」グラフを確認します。 ピークの「アクティブな」インスタンス値は、ビジュアライザーグラフに表示された値とほぼ一致する必要があります。 グラフが更新されるまで約 3 分待つ必要があります。

>Note: クラウドコンソールの[指標]タブには、Cloud Run サービスに関する最も正確な情報が表示されます。 この情報は、Cloud Monitoring から取得されます。 ただし、コンソールのメトリックは更新に約 3 分かかります。 このラボでは、visualizer グラフを使用してリアルタイムのスケーリングを示します。 visualizer はデモ専用です。

### サービスのコンカレンシーをアップデート

Cloud Run は、特定のコンテナインスタンスで同時に処理できるリクエストの最大数を指定する [concurrency](https://cloud.google.com/run/docs/about-concurrency) 設定を提供します。

コードで並列リクエストを処理できない場合は、 `concurrency=1` を設定してください。図のように、各コンテナインスタンスは一度に 1 つのリクエストのみを処理します。

コンテナが複数のリクエストを同時に処理できる場合は、より高いコンカレンシーを設定します。指定されたコンカレンシー値は _maximum_ であり、インスタンスの CPU がすでに高度に使用されている場合、Cloud Run は特定のコンテナインスタンスに対してそれほど多くの要求を費やさない可能性があります。図では、サービスは最大 80 の同時要求（デフォルト）を処理するように構成されています。したがって、Cloud Run は、3 つのリクエストすべてを単一のコンテナインスタンスに送信します。

![](image/concurrency_image.png?raw=true)

`concurrency=1`の初期設定で metrics-writer サービスをデプロイしました。これは、各コンテナインスタンスが一度に 1 つのリクエストのみを処理することを意味します。この値を使用して、Cloud Run の高速自動スケーリングを示しました。ただし、このような単純なサービスでは、おそらくはるかに高い同時実行性を処理できます。ここでは、コンカレンシー設定を増やして、スケーリング動作への影響を調査します。

1. metrics-writer サービスのコンカレンシー設定を更新します。これにより、サービスの新しいリビジョンが作成されます。準備が整うと、すべてのリクエストがこの新しいリビジョンにルーティングされます。

```bash
gcloud run services update metrics-writer \
  --concurrency 5
```

2. コマンドを再実行して、負荷を生成します。
```bash
hey -z 30s -c 30 $WRITER_URL
```

3. visualizer Web アプリを表示するブラウザーページに戻ります。ページに別のグラフがプロットされています。

![](image/visualizer_graph_3.png?raw=true)

4. `hey` 出力の要約を確認します。

![](image/hey_summary_2.png?raw=true)

### サービスの最大インスタンス構成を更新する

ここでは、[コンテナ インスタンスの最大数](https://cloud.google.com/run/docs/about-instance-autoscaling#max-instances) 設定を使用して、リクエストに応じたサービスのスケーリングを制限します。 この設定は、コストを管理したり、データベースなどのバックエンドサービスへの接続数を制限したりする方法として使用します。

1. metrics-writer サービスの最大インスタンス設定を更新します。
```bash
gcloud run services update metrics-writer \
  --max-instances 5
```

2. コマンドを再実行して、負荷を生成します。
```bash
hey -z 30s -c 30 $WRITER_URL
```

3. visualizer Web アプリを表示するブラウザーページに戻ります。ページに別のグラフがプロットされています。

4. `hey` 出力の要約を確認します。以前の出力と比較してどうですか？

## 3. 軽快なトラフィック

>_**クラウドネイティブの原則**: クラウドネイティブアプリには、プログラム可能なネットワークデータプレーンがあります。_

このセクションでは、Cloud Run トラフィックの分割と ingress ルールを構成します。 このネットワーク動作は、シンプルな API 呼び出しを使用してプログラムします。

![](image/nimble-traffic_image.png?raw=true)

### タグ付きバージョンをデプロイする

名前付きタグを新しいリビジョンに割り当てることができます。これにより、サービストラフィックなしで、特定の URL でリビジョンにアクセスできます。次に、そのタグを使用して、トラフィックをタグ付きリビジョンに徐々に移行し、タグ付きリビジョンをロールバックできます。この機能の一般的な使用例は、トラフィックを処理する前に、新しいサービスリビジョンのテストと検証に使用することです。

1. Cloud Shell を開きます。以前のシェルがしばらく非アクティブだった場合は、再接続が必要になる場合があります。その場合は、再接続後、repo ディレクトリに移動し、環境変数を再設定します。
```bash
cd ~/cloudshell_open/gig-training-materials/gig08-01/ && source vars.sh
```

2. metrics-writer サービスの新しいリビジョンをデプロイし、コンカレンシーと最大インスタンス値を既知の値に戻します。
```bash
gcloud run services update metrics-writer \
  --concurrency 5 \
  --max-instances 7
```

3. metrics-writer サービスの新しいリビジョンをデプロイします。 'green' というタグを指定します。 `--no-traffic` フラグを設定します。これは、トラフィックが新しいリビジョンにルーティングされないことを意味します。表示されるグラフの色を制御する LABEL 環境変数を設定します（環境変数はタグとはまったく関係がないことに注意してください）。
```bash
gcloud run deploy metrics-writer \
  --tag green \
  --no-traffic \
  --set-env-vars LABEL=green \
  --image asia-northeast1-docker.pkg.dev/gig6-1/gig6-1/metrics-writer:latest
```

以下のような出力が表示されます。リビジョンはトラフィックの 0％ を処理しており、タグ名の前に専用 URL が付いていることに注意してください。

**Output**
```terminal
Deploying container to Cloud Run service [metrics-writer] in project [gig7-1] region [asia-northeast1]
OK Deploying... Done.
  OK Creating Revision...
  OK Routing traffic...
Done.
Service [metrics-writer] revision [metrics-writer-00005-don] has been deployed and is serving 0 percent of traffic.
The revision can be reached directly at https://green---metrics-writer-rmclwajz3a-an.a.run.app
```

4. 前のコマンドは、新しい 'green' リビジョン専用のタグ付き URL を出力します。 [TAGGED_URL] をコマンド出力の値に置き換えて、シェル変数を設定します。
```bash
export GREEN_URL=[TAGGED_URL]
```

5. サービスリビジョンを一覧表示します。 2 つのアクティブなリビジョンがあることがわかります。
```bash
gcloud run revisions list --service metrics-writer
```

6. サービスへのリクエストを実行します。コマンドを数回実行すると、サービスが常に 'blue' を返すことがわかります。'green' サービスは、プライマリサービス URL からのトラフィックを処理していません。
```bash
curl $WRITER_URL
```

7. 新しいタグ付き URL に対してリクエストを実行します。サービスが 'green' を返すことがわかります。メインリビジョンは引き続きすべてのライブトラフィックに対応していますが、テスト可能な専用 URL が付いたタグ付きバージョンがあります。
```bash
curl $GREEN_URL
```

**Output**
```terminal
Hello from green
```

### トラフィックスプリッティングの構成

Cloud Run を使用すると、トラフィックを受信するリビジョンまたはタグを指定したり、リビジョンによって受信されるトラフィックの割合を指定したりできます。この機能を使用すると、以前のリビジョンにロールバックし、リビジョンを段階的にロールアウトして、トラフィックを複数のリビジョンに分割できます。

1. トラフィック分割を構成し、トラフィックの 10％　を 'green' とタグ付けされたリビジョンに送信します。
```bash
gcloud run services update-traffic \
  metrics-writer --to-tags green=10
```

以下のような出力が表示されます。出力には、現在のトラフィック構成が記述されています。

**Output**
```terminal
OK Updating traffic... Done.
  OK Routing traffic...
Done.
URL: https://metrics-writer-rmclwajz3a-an.a.run.app
Traffic:
  90% metrics-writer-00004-wof
  10% metrics-writer-00005-don
        green: https://green---metrics-writer-rmclwajz3a-an.a.run.app
```

2. metrics-writer サービスに対してリクエストロードを生成します。メインサービスの URL が表示されます。
```bash
hey -z 30s -c 30 $WRITER_URL
```

3. visualizer Web アプリを表示するブラウザーページに戻ります。ページにグラフがプロットされています。グラフには、緑と青の2本の線があります。 'green' サービスはトラフィックの約 10％ を受信して​​います。

![](image/visualizer_graph_4.png?raw=true)

4. 別のトラフィック分割を構成し、トラフィックの 50％ を 'green' とタグ付けされたリビジョンに送信します。
```bash
gcloud run services update-traffic \
  metrics-writer --to-tags green=50
```

5. metrics-writer サービスに対してリクエストロードを生成します。
```bash
hey -z 30s -c 30 $WRITER_URL
```

6. visualizer Web アプリを表示するブラウザーページに戻ります。今回は、トラフィックが 'green' と 'blue' のリビジョン間で均等に分割されます。

### 外部 HTTP(S) ロードバランサーを作成する

このセクションでは、[外部 HTTP(S) 負荷分散の概要](https://cloud.google.com/load-balancing/docs/https) を作成します。 Google Cloud HTTP（S）Load Balancing は、グローバルなプロキシベースのレイヤー 7 ロードバランサーであり、単一の外部 IP アドレスの背後でサービスを世界中で実行および拡張できます。 [サーバーレス ネットワーク エンドポイント グループ](https://cloud.google.com/load-balancing/docs/negs/serverless-neg-concepts)(NEG) を使用して、ロードバランサーから Cloud Run サービスにリクエストをルーティングします。

>WARNING: 簡単にするために、HTTP（HTTPSではない）ロードバランサーを作成します。この方法では、証明書を設定する必要はありません。本番環境では、HTTPS ロードバランサーを使用する必要があります。

>NOTE: ロードバランサーの作成にはいくつかの手順が必要です。ここでは Terraform を使用してロードバランサーと関連コンポーネントを作成します。

1. gig08-01 リポジトリディレクトリで、Terraform を初期化します。
```bash
terraform init
```

2. Terraform で構成を適用して、ロードバランサーと関連コンポーネントを作成します。ロードバランサーから metrics-writer の Cloud Run サービスにリクエストをルーティングするサーバーレス NEG を構成します。
```bash
terraform apply -auto-approve -var project_id=$PROJECT_ID
```

Terraform 出力の最終行は、次のようになります:

**Output**
```terminal
Apply complete! Resources: 7 added, 0 changed, 0 destroyed.
```

3. ロードバランサーの外部 IP アドレスを指定する転送ルールを一覧表示します。
```bash
gcloud compute forwarding-rules list
```

以下のような出力が表示されます。

**Output**
```terminal
NAME: lb-http
REGION:
IP_ADDRESS: 34.110.187.86
IP_PROTOCOL: TCP
TARGET: lb-http-http-proxy
```

4. ロードバランサーの IP アドレスのシェル変数を設定し、[IP_ADDRESS] を前の出力の値に置き換えます。
```bash
export LB_IP=[IP_ADDRESS]
```

5. ロードバランサーが完全にオンラインになるまで **1分** 待ちます。

6. ロードバランサーのアドレスに対して HTTPGET 要求を実行します。 404 エラーが発生した場合は、ロードバランサーの準備ができるまでもう少し待ちます。
```bash
curl $LB_IP
```

metrics-writer サービスからの応答が表示されます。ロードバランサーへの HTTP リクエストは、metrics-writer サービスにルーティングされています。トラフィック分割がまだアクティブであるため、'green' または 'blue' のいずれかが表示される場合があります。

**Output**
```terminal
Hello from blue
```

### ingress ルールを適用する

`--allow-unauthenticated` フラグを使用して metrics-writer Cloud Run サービスをデプロイしました。これにより、サービス URL がインターネット上で公開され、誰でもあなたのサービスに接続することができます。

このセクションでは、Cloud Run サービスで [上り（内向き）の設定](https://cloud.google.com/run/docs/securing/ingress) を設定して、ロードバランサーまたプロジェクトの VPC ネットワーク内部から発信されていないリクエストを拒否します。これにより、Cloud Run サービスの URL はインターネット上でパブリックにアクセスできません。

ロードバランサーを介してすべてのリクエストを強制することで、[Cloud Armor](https://cloud.google.com/armor) や [Cloud CDN](https://cloud.google.com/cdn) などの追加のロードバランサー機能を利用することもできます。

1. サービス URL を介して metrics-writer サービスに引き続き接続できることを確認します。
```bash
curl $WRITER_URL
```

2. ingress ルールを metrics-writer Cloud Run サービスに適用します。ingress ルールでは、Google Cloud ロードバランサーまたはプロジェクトの VPC 内から発信されたリクエストのみが許可されます。
```bash
gcloud run services update metrics-writer \
  --ingress internal-and-cloud-load-balancing
```

3. サービスの URL を介してサービスを操作できなくなったことを確認します。
```bash
curl $WRITER_URL
```

サービス URL へのリクエストは拒否されます。 HTTP 403(禁止) エラーを説明する HTML ページが表示されます。インターネット上でサービス URL にアクセスできなくなりました。

**Output**
```terminal
<html><head>
<meta http-equiv="content-type" content="text/html;charset=utf-8">
<title>404 Not Found</title>
</head>
<body text=#000000 bgcolor=#ffffff>
<h1>Error: Not Found</h1>
<h2>The requested URL <code>/</code> was not found on this server.</h2>
<h2></h2>
</body></html>
```

4. ロードバランサーを介してサービスに接続できることを確認します。
```bash
curl $LB_IP
```

## おめでとうございます!

<walkthrough-conclusion-trophy></walkthrough-conclusion-trophy>

これでこのラボは完了です。
- Google Cloud のサーバーレスコンテナプラットフォームである Cloud Run を使用して、いくつかの主要なクラウドネイティブの原則を実装しました。
- コンテナ化された Web サービスを展開し、高速自動スケーリングをトリガーし、ネットワークトラフィックを操作し、組み込みのセキュリティ制御を適用しました。
- Cloud Run を使用して、クラウドネイティブアプリケーションのモダナイゼーションを加速できることを学びました。

デモで使った資材が不要な方は、次の手順でクリーンアップを行って下さい。

## **クリーンアップ（プロジェクトを削除）**

ハンズオン用に利用したプロジェクトを削除し、コストがかからないようにします。

### **1. Google Cloud のデフォルトプロジェクト設定の削除**

```bash
gcloud config unset project
```

### **2. プロジェクトの削除**

```bash
gcloud projects delete $PROJECT_ID
```

### **3. ハンズオン資材の削除**

```bash
cd $HOME && rm -rf ./cloudshell_open/gig-training-materials
```
# Additional Contents
以降に、Cloud Run に関する追加のコンテンツを 2 つ用意しています。ハンズオンの時間が余った方や、もっとサーバーレスの勉強をしたいという方はぜひ挑戦してみてください。

## 4. [セキュリティのシフトレフト](https://cloud.google.com/architecture/devops/devops-tech-shifting-left-on-security)

>_**クラウドネイティブの原則**：クラウドネイティブアプリはセキュリティに早期に対応し、プラットフォームのセキュリティ機能を使用します。_
>Note: 「シフトレフト」とは、セキュリティに関する問題がソフトウェア開発ライフサイクルの早い段階で対処されることを意味します（左から右のスケジュール図で左側に位置します）。

このタスクでは、新しい ID を metrics-writer サービスに割り当て、制限付きのアクセス許可を付与します。さらに、metrics-writer Cloud Run サービスへのユーザーアクセスを制限します。

### 専用のサービス ID を割り当てる

Cloud Run リビジョンは、 [サービスアカウント](https://cloud.google.com/iam/docs/service-accounts) を [ランタイム ID](https://cloud.google.com/run/docs/securing/service-identity) として使用します。サービスアカウントは、個人ではなく、アプリケーションによって使用される特別な種類のアカウントです。アプリケーションはサービスアカウントを使用して[許可されたAPI呼び出し](https://developers.google.com/identity/protocols/oauth2/service-account#authorizingrequests) を行います。

コードが [Cloud クライアント ライブラリ](https://cloud.google.com/apis/docs/cloud-client-libraries) を使用して Google Cloud API と対話する場合、ランタイムサービスアカウントから資格情報を自動的に取得して使用します。この戦略は ["アプリケーションデフォルトクレデンシャル"](https://cloud.google.com/docs/authentication/production#providing_credentials_to_your_application) と呼ばれます。

次の図は、metrics-writer インスタンスが Firestore に書き込むときに使用されるアプリケーションのデフォルトの資格情報アプローチを示しています。クライアントライブラリは、ランタイムサービスアカウントの ID トークンを自動的にフェッチし、それを Firestore リクエストに添付します。 Firestore API はリクエストを認証および承認し、サービスアカウントに Firestore への書き込みに適切な IAM 権限があることを確認します。

![](image/security_image.png?raw=true)

>デフォルトでは、Cloud Runリビジョンは Compute Engine のデフォルトサービスアカウント `PROJECT_NUMBER-compute@developer.gserviceaccount.com` を使用します。このアカウントには、プロジェクト[編集者](https://cloud.google.com/iam/docs/understanding-roles#basic) IAM ロールがあります。これは、デフォルトで、Cloud Run リビジョンが Google Cloud プロジェクトのすべてのリソースへの読み取りおよび書き込みアクセス権を持っていることを意味します。

デフォルトのサービスアカウントを使用する代わりに、ユーザーが管理するサービスアカウントを割り当てることにより、各サービスに[専用 ID](https://cloud.google.com/run/docs/securing/service-identity#per-service-identity) を付与することをお勧めします。ユーザー管理のサービスアカウントを使用すると、[IAM を使用して](https://cloud.google.com/iam/docs/creating-managing-service-accounts) 最小限の権限セットを付与することでアクセスを制御できます。

1. Cloud Shell を開きます。以前のシェルがしばらく非アクティブだった場合は、再接続が必要になる場合があります。その場合は、再接続後、repo ディレクトリに移動し、環境変数を再設定します。
```bash
cd ~/cloudshell_open/gig-training-materials/gig08-01/ && source vars.sh
```

2. 新しいサービスアカウントを作成します。
```bash
gcloud iam service-accounts create metrics-writer-sa \
  --description "Runtime service account for metrics-writer service"
```

3. 新しいサービスアカウントを metrics-writer サービスに割り当てます。これにより、新しい metrics-writer リビジョンが作成されます。
```bash
gcloud run services update metrics-writer \
  --service-account metrics-writer-sa@$PROJECT_ID.iam.gserviceaccount.com
```

**Output**
```terminal
OK Deploying... Done.
  OK Creating Revision...
Done.
Service [metrics-writer] revision [metrics-writer-00009-vos] has been deployed and is serving 0 percent of traffic.
```

4. トラフィックの 100％ を新しいリビジョンにルーティングします。 
```bash
gcloud run services update-traffic metrics-writer --to-latest
```

5. ロードバランサーを介してサービスを呼び出し、詳細を出力します。
```bash
curl -v $LB_IP
```

HTTP503エラーと "Instance not ready" というメッセージが表示されます。

**Output**
```terminal
*   Trying 34.110.187.86:80...
* Connected to 34.110.187.86 (34.110.187.86) port 80 (#0)
> GET / HTTP/1.1
> Host: 34.110.187.86
> User-Agent: curl/7.74.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 503 Service Unavailable
< content-type: text/plain; charset=utf-8
< etag: W/"13-E9/U2/cQlxfYtpGiFJun9Uohjjc"
< X-Cloud-Trace-Context: e3999db428f1de8a53d37e1b3b27424c;o=1
< Date: Fri, 10 Jun 2022 04:47:26 GMT
< Server: Google Frontend
< Content-Length: 19
< Via: 1.1 google
<
Instance not ready
* Connection #0 to host 34.110.187.86 left intact
```

6. クラウドコンソールの [Cloud Run セクション](https://console.cloud.google.com/run) にアクセスします。metrics-writer サービスをクリックし、'ログ'タブを選択します。

![](image/metrics-writer_logs.png?raw=true)

PERMISSION_DENIEDエラーが表示されます。 エラートレースを見ると、エラーが Firestore に関連していることがわかります。 metrics-writer リビジョンのランタイムサービスアカウントとして割り当てた新しいサービスアカウントには、Firestore に書き込むための適切な権限がありません。

7. Cloud Shell に戻り、metrics-writer ランタイム ID として使用されるサービスアカウントに適切な Firestore IAM ロールを付与します。
```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --role roles/datastore.user \
  --member "serviceAccount:metrics-writer-sa@$PROJECT_ID.iam.gserviceaccount.com"
```

更新された IAM ポリシーが表示されます。

**Output**
```terminal
Updated IAM policy for project [gig7-1].
bindings:
...
- members:
  - serviceAccount:metrics-writer-sa@gig7-1.iam.gserviceaccount.com
  role: roles/datastore.user
...
```

8. IAM 権限が伝播するまで**1分**待ちます。

9. サービスを再度呼び出します。 正常な応答を受け取ります。
```bash
curl $LB_IP
```
専用の ID を metrics-writer サービスに割り当て、その ID に、正しく動作するために必要な制限された権限のみを付与しました。

### 追加の認証

`--allow-unauthenticated` フラグを使用して metrics-writer Cloud Run サービスをデプロイしました。このフラグにより​​、サービス URL がインターネット上で公開されます。また、[上り（内向き）の制限](https://cloud.google.com/run/docs/securing/ingress) を設定して、Google Cloud ロードバランサーからのリクエストのみを受け入れるようにし、metrics-writer サービスのパブリック URL を効果的に無効にします。

しかし、このサービスはまだロードバランサーの IP アドレスを介してパブリックにアクセスできます。誰でも metrics-writer サービスを呼び出すことができます。

このセクションでは、Cloud Run の組み込みの認証および承認機能を使用して、既知の関係者からの要求のみを受け入れます。サービスの認証を有効にすると、そのサービスに対して[Cloud Run Invoker IAM ロール](https://cloud.google.com/run/docs/reference/iam/roles)(`roles / run.invoker`) が明示的に付与されている ID のみがサービスを呼び出すことができます。

>Note: 通常、Cloud Run Invoker の役割を使用して、サービスアカウント、または内部ユーザーやグループに権限を付与します。 Web またはモバイルアプリでエンドユーザーを認証するためには、別のアプローチを使用します。詳細については、[認証の概要](https://cloud.google.com/run/docs/authenticating/overview)のドキュメントを参照してください。

1. metrics-writer の IAM ポリシーを確認します。
```bash
gcloud run services get-iam-policy metrics-writer
```


`allUsers` ID に`run.invoker` ロールがあることがわかります。これは、誰でもサービス(パブリック)を呼び出すことができることを意味します。この `allUsers` IAM ポリシーは、`--allow-unauthenticated` フラグを使用するとサービスに追加されます。

**Output**
```terminal
bindings:
- members:
  - allUsers
  role: roles/run.invoker
etag: BwXg-lSl5iI=
version: 1
```

2. metrics-writer サービスの新しいリビジョンをデプロイします。 `--no-allow-unauthenticated` フラグを指定します。
```bash
gcloud run deploy metrics-writer \
  --no-allow-unauthenticated \
  --image asia-northeast1-docker.pkg.dev/gig6-1/gig6-1/metrics-writer:latest
```

3. metrics-writer の IAM ポリシーを確認します。
```bash
gcloud run services get-iam-policy metrics-writer
```

`allUsers` IAM ポリシーが存在しなくなったことがわかります。

**Output**
```terminal
etag: BwXhEK-IySg=
version: 1
```

4. ロードバランサーを介して metrics-writer サービスを呼び出します。
```bash
curl $LB_IP
```

HTTP 403 forbedden エラーが表示されます。これで、認証および承認されたユーザーのみがサービスを呼び出すことができます。

**Output**
```terminal
<html><head>
<meta http-equiv="content-type" content="text/html;charset=utf-8">
<title>403 Forbidden</title>
</head>
<body text=#000000 bgcolor=#ffffff>
<h1>Error: Forbidden</h1>
<h2>Your client does not have permission to get URL <code>/</code> from this server.</h2>
<h2></h2>
</body></html>
```

5. ログインしているユーザーのシェル変数を設定します。
```bash
export USER=$(gcloud config get-value account); echo $USER
```

**Output**
```terminal
Your active configuration is: [cloudshell-15211]
somebody@somedomain.com
```

6. IAM 呼び出し側の役割を、metrics-writer サービスのユーザーに付与します。
```bash
gcloud run services add-iam-policy-binding metrics-writer \
  --role roles/run.invoker --member "user:$USER"
```

新しいIAMポリシーが出力されます。

**Output**
```terminal
Updated IAM policy for service [metrics-writer].
bindings:
- members:
  - user:somebody@somedomain.com
  role: roles/run.invoker
etag: BwXhELhW-90=
version: 1
```

7. ロードバランサーを介して metrics-writer サービスを呼び出します。
```bash
curl $LB_IP
```

まだ HTTP 403 forbidden エラーが表示されます。これは、`curl` を介して送信したリクエストに ID 情報が添付されていないためです。 Cloud Run はリクエストを認証できません。

8. metrics-writer サービスを再度呼び出しますが、今回は要求ヘッダーに ID トークンを含めます。 gcloud を使用して、Cloud Shell にログインしているユーザーの ID トークンを生成します。ロードバランサーは ID トークンを Cloud Run に渡します。
```bash
curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" $LB_IP
```

ID トークンを使用すると、Cloud Run はリクエストを行うユーザー ID を認証し、ユーザーがサービスを呼び出すための適切な `run.invoker` 権限を持っていることを確認できます。呼び出しは成功します。ラベルの出力が表示されます。
**Output**
```terminal
Hello from green
```

## 5. AI/ML との連携

このセクションでは、Cloud Run と Cloud Storage、そして Cloud Vision を組み合わせたサーバーレス システムを構築します。Google Cloud では、事前トレーニング済みの機械学習モデルの API が提供されているため、Web アプリケーションと AI/ML の連携を容易に実現できます。

![](image/aiml_image.png?raw=true)

AI/ML web app
- 画像データのアップロードが可能なシンプルな Web アプリケーション。
- 画像データがアップロードされると、Cloud Vision API と連携して画像データのラベルを推定し、Web サイトに表示します。

Cloud Storage
- ユーザーからアップロードされた画像データを格納します。

Cloud Vision
- Cloud Storage に格納された画像データを読み込み、ラベルを推定します。

### 事前準備
1. Cloud Shell を開きます。以前のシェルがしばらく非アクティブだった場合は、再接続が必要になる場合があります。その場合は、再接続後、リポディレクトリに移動し、環境変数を再設定します。

```bash
cd ~/cloudshell_open/gig-training-materials/gig08-01/ && source vars.sh
```

2. 必要な API を有効にします。

```bash
gcloud services enable vision.googleapis.com \
  storage-component.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com
```

3. Cloud Shell の右上の「エディタを開く」をクリックし、左のエクスプローラーから gig08-01 ディレクトリ内の更に aiml_cloudrun ディレクトリ内にある main.py を選択します。

![](image/source_code.png?raw=true)

4. main.py の以下の値部分を、ご自身のプロジェクト ID に変更します。

```
bucket_name = '<プロジェクト ID をここに設定>'
```

なお、プロジェクト ID は Cloud Shell のターミナルから以下のコマンドを入力することで確認できます。

```bash
gcloud config get-value project
```

5. GCP のプロジェクト番号を環境変数に設定します。

```bash
export PROJECT_NUM=$(gcloud projects describe $PROJECT_ID --format json | jq -r '.projectNumber')
```

6. Cloud Build が使用するサービスアカウントに、Cloud Storage に対する読み込み権限を付与します。

```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --role roles/storage.objectViewer \
  --member "serviceAccount:$PROJECT_NUM@cloudbuild.gserviceaccount.com"
```

**Output**
```terminal
Updated IAM policy for project [gig7-1].
bindings:
- members:
  - serviceAccount:XXX@cloudbuild.gserviceaccount.com
  role: roles/cloudbuild.builds.builder
- members:
  - serviceAccount:service-XXX@gcp-sa-cloudbuild.iam.gserviceaccount.com
  role: roles/cloudbuild.serviceAgent
- members:
  - serviceAccount:service-XXX@compute-system.iam.gserviceaccount.com
  role: roles/compute.serviceAgent
- members:
  - serviceAccount:service-XXX@containerregistry.iam.gserviceaccount.com
  role: roles/containerregistry.ServiceAgent
- members:
  - serviceAccount:XXX@cloudservices.gserviceaccount.com
  role: roles/editor
- members:
  - serviceAccount:service-XXX@firebase-rules.iam.gserviceaccount.com
  role: roles/firebaserules.system
- members:
  - user:XXX
  role: roles/owner
- members:
  - serviceAccount:service-XXX@gcp-sa-pubsub.iam.gserviceaccount.com
  role: roles/pubsub.serviceAgent
- members:
  - serviceAccount:service-XXX@serverless-robot-prod.iam.gserviceaccount.com
  role: roles/run.serviceAgent
- members:
  - serviceAccount:XXX@cloudbuild.gserviceaccount.com
  role: roles/storage.objectViewer
etag: BwX3eTaMd34=
version: 1
```

### 画像データ格納用のバケット作成
1. Cloud Storage バケットを作成します。

**※このバケットに格納されるオブジェクトは、世界中のユーザーから参照が可能となります。取り扱いに十分ご注意ください。**

```bash
gcloud storage buckets create gs://$PROJECT_ID \
  --location=asia-northeast1 \
  --uniform-bucket-level-access \
  --no-public-access-prevention
```

2. バケット内のすべてのオブジェクトを公開するように設定します。

```bash
gcloud storage buckets add-iam-policy-binding gs://$PROJECT_ID \
  --member=allUsers \
  --role=roles/storage.objectViewer
```

**Output**
```terminal
bindings:
- members:
  - projectEditor:gig7-1
  - projectOwner:gig7-1
  role: roles/storage.legacyBucketOwner
- members:
  - projectViewer:gig7-1
  role: roles/storage.legacyBucketReader
- members:
  - projectEditor:gig7-1
  - projectOwner:gig7-1
  role: roles/storage.legacyObjectOwner
- members:
  - projectViewer:gig7-1
  role: roles/storage.legacyObjectReader
- members:
  - allUsers
  role: roles/storage.objectViewer
etag: CAI=
kind: storage#policy
resourceId: projects/_/buckets/<bucket name>
version: 1
```

### AI/ML web app のデプロイ
1. Cloud Run でアプリケーションをデプロイします。今回はコンテナイメージからではなく、ソースコードからデプロイします。この場合、Artifact Registry にビルドされたコンテナイメージが保存されます。

```bash
gcloud run deploy aiml-web-application \
  --allow-unauthenticated \
  --source ~/cloudshell_open/gig-training-materials/gig08-01/aiml_cloudrun
```

**Output**
```terminal
Deploying from source requires an Artifact Registry Docker repository to store built containers. A repository named [cloud-run-source-deploy] in region [asia-northeast1] will be created.

Do you want to continue (Y/n)?  y

This command is equivalent to running `gcloud builds submit --tag [IMAGE] ~/cloudshell_open/gig-training-materials/gig08-01/aiml_cloudrun --image [IMAGE]`

Building using Dockerfile and deploying container to Cloud Run service [aiml-web-application] in project [gig7-1] region [asia-northeast1]
OK Building and deploying new service... Done.                                                                  
  OK Creating Container Repository...
  OK Uploading sources...
  OK Building Container... Logs are available at [https://console.cloud.google.com/cloud-build/builds/8c0e3b85-1804-4b52-908f-29260c4a3647?project=156231959753].
  OK Creating Revision... Creating Service.
  OK Routing traffic...
  OK Setting IAM Policy...
Done.
Service [aiml-web-application] revision [aiml-web-application-00001-fuh] has been deployed and is serving 100 percent of traffic.
Service URL: https://aiml-web-application-4jkzyhm2aq-an.a.run.app
```

2. 以下のコマンドで AI/ML web app アプリケーションの URL を確認し、ブラウザからアクセスします。

```bash
gcloud run services describe aiml-web-application --format='value(status.url)'
```

3. 任意の画像データをアップロードし、挙動を確認してみてください。

4. Cloud Storage に画像データが格納されていることを確認します。

```bash
gcloud storage ls gs://$PROJECT_ID
```

## お疲れ様でした!

<walkthrough-conclusion-trophy></walkthrough-conclusion-trophy>

これで追加のコンテンツは完了です。
- Cloud Run でデプロイしたアプリケーションに関して、サービスアカウントによるアクセス制御の挙動を確認しました。
- Cloud Run と Cloud Vision を使用して、AI/ML と連携するサーバーレスの Web アプリケーションをデプロイしました。

再掲となりますが、デモで使った資材が不要な方は、次の手順でクリーンアップを行って下さい。

## **クリーンアップ（プロジェクトを削除）**

ハンズオン用に利用したプロジェクトを削除し、コストがかからないようにします。

### **1. Google Cloud のデフォルトプロジェクト設定の削除**

```bash
gcloud config unset project
```

### **2. プロジェクトの削除**

```bash
gcloud projects delete $PROJECT_ID
```

### **3. ハンズオン資材の削除**

```bash
cd $HOME && rm -rf ./cloudshell_open/gig-training-materials
```