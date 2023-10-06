# G.I.G. 07-02 データベースセクションハンズオン

## 環境準備

### Google Cloud プロジェクトの選択

ハンズオンを行う Google Cloud プロジェクトをまだ作成されていない場合は、[こちらのリンク](https://console.cloud.google.com/projectcreate) から新しいプロジェクトを作成してください。

**なるべく新しいプロジェクトが望ましいです。**

それでは最初に、ハンズオンを進めるための環境準備を行います。

#### GCP のプロジェクト ID を環境変数に設定

環境変数 `GOOGLE_CLOUD_PROJECT` に GCP プロジェクト ID を設定します。下のプルダウンメニューから、ご使用になられる Google Cloud プロジェクトを選択して下さい。

<walkthrough-project-setup></walkthrough-project-setup>

```bash
export GOOGLE_CLOUD_PROJECT=<walkthrough-project-id/>
```
※ もし上のプルダウンメニューが正しく動作しない場合、以下のコマンドを実行して下さい。 `[PROJECT_ID]` はご自身のプロジェクト ID に置き換えてください。

```bash
export GOOGLE_CLOUD_PROJECT=[PROJECT_ID]
```

#### CLI（gcloud コマンド）から利用する GCP のデフォルトプロジェクトを設定

操作対象のプロジェクトを設定します。

```bash
gcloud config set project $GOOGLE_CLOUD_PROJECT
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
cd ~/cloudshell_open/gig-training-materials/gig07-02/
```

### **2. チュートリアルを開く**

```bash
teachme tutorial.md
```

---

## Cloud Run 上の Node.js アプリケーションを Cloud SQL for PostgreSQL データベースに接続する

### 1. **概要**

[Cloud SQL Node.js コネクタ](https://github.com/GoogleCloudPlatform/cloud-sql-nodejs-connector#readme) は、 Node.js アプリケーションを Cloud SQL データベースに接続する最も簡単かつセキュアに接続する方法です。また、 [Cloud Run](https://cloud.google.com/run) は、 HTTP リクエストで呼び出すことが出来、ステートレスなコンテナを動かすことを可能にするフルマネージドなサーバーレスプラットフォームです。このラボでは、 Cloud Run 上の Node.js アプリケーションを Cloud SQL for PostgreSQL にサービスアカウントと IAM 認証を使ってセキュアに接続する方法を試します。

#### **目的**

このハンズオンでは、次の内容を実行します。

- Cloud SQL の PostgreSQL インスタンスを作成します。
- Node.js アプリケーションを Cloud Run にデプロイします。
- Cloud SQL Node.js コネクタライブラリを使ってアプリケーションをデータベースに接続します。

#### **前提条件**

このハンズオンは、Cloud Console および Cloud Shell 環境の理解を前提としています。

## 2. 始める前に

### 環境のセットアップ

1. `Cloud Shell` を開きます。

>Note: README の青い`OPEN IN GOOGLE CLOUD SHELL` ボタンから開始された場合は、すでにリポジトリはクローンされていますので、4 にスキップしてください。

2. このハンズオンのスクリプトを含む git リポジトリをクローンします。 gcloud の承認を求められた場合は、承認してください。
```bash
git clone https://github.com/google-cloud-japan/gig-training-materials.git
```

3. リポジトリディレクトリに移動します。
```bash
cd gig07-02
```

4. Cloud Shell から API を有効にします。:

```bash
gcloud services enable compute.googleapis.com sqladmin.googleapis.com \
  run.googleapis.com artifactregistry.googleapis.com \
  cloudbuild.googleapis.com servicenetworking.googleapis.com
```

このコマンドは完了するまでに数分かかる場合がありますが、最終的には次のような成功メッセージが表示されるはずです。:

```
Operation "operations/acf.p2-327036483151-73d90d00-47ee-447a-b600-a6badf0eceae" finished successfully.
```

## 3. **サービスアカウントのセットアップ**

Cloud Run が使用するサービスアカウントを作成し、Cloud SQL への正しいアクセス権を割り当てます。

1. `gcloud iam service-accounts create` コマンドを実行して新しいサービスアカウントを作成します。:

```bash
gcloud iam service-accounts create quickstart-service-account \
  --display-name="Quickstart Service Account"
```

1. `gcloud projects add-iam-policy-binding` コマンドを実行して、先ほど作成したサービスアカウントに Cloud SQL クライアントの権限を割り当てます。

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member="serviceAccount:quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"
```

3. `gcloud projects add-iam-policy-binding` コマンドを実行して、先ほど作成したサービスアカウントに Cloud SQL インスタンス ユーザーの権限を割り当てます。

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member="serviceAccount:quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --role="roles/cloudsql.instanceUser"

```

4. `gcloud projects add-iam-policy-binding` コマンドを実行して、先ほど作成したサービスアカウントに Cloud Logging ログ書き込み権限を割り当てます。

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member="serviceAccount:quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --role="roles/logging.logWriter"
```

## 4. **Cloud SQL のセットアップ**
`gcloud sql instances create` コマンドを実行して、Cloud SQL インスタンスを作成します。

- **-database-version**: データベースエンジンのタイプとバージョンの指定。指定されない場合は API のデフォルト値が使用されます。詳しくは gcloud データベース バージョンに関する [ドキュメント] (https://cloud.google.com/sql/docs/db-versions?hl=ja) に記載されている現在利用可能なバージョンをご確認下さい。
- **-cpu**: インスタンスに必要とされる CPU コアの数
- **-memory**: インスタンスに必要とされるメモリ容量 (例: 3072MB, 9GB)
- **–region**: インスタンスを配置するリージョン (例: us-central1, asia-northeast1, us-east1)
- **–database-flags**: データベースエンジン固有のパラメータであるフラグの設定。このケースでは、Cloud Run から Cloud SQL へのサービスアカウントによる接続を許可するために  `cloudsql.iam_authentication` というフラグを On にします。

```bash
gcloud sql instances create quickstart-instance \
  --database-version=POSTGRES_14 \
  --cpu=1 \
  --memory=4GB \
  --region=asia-northeast1 \
  --database-flags=cloudsql.iam_authentication=on
```

このコマンドの完了には数分掛かります。

`gcloud sql databases create` コマンドを実行して、 `quickstart-instance` の中に Cloud SQL データベースを作成します。

```bash
gcloud sql databases create quickstart_db \
  --instance=quickstart-instance
```

サービスアカウントを使ってデータベースにアクセスさせるため、先に作成したサービスアカウントと同名の PostgreSQL データベース ユーザーを作成します。

```bash
gcloud sql users create quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam \
  --instance=quickstart-instance \
  --type=cloud_iam_service_account
```

## 5. **アプリケーションの準備**

HTTP リクエストに応答する Node.js アプリケーションを準備します。アプリケーションは `gig07-02/helloworld` というディレクトリにあります。

1. Cloud Shell で `helloworld` ディレクトリに移動します:

```bash
cd helloworld
```

2. `package.json` ファイルをモジュールとして初期化します。

```bash
npm init -y
npm pkg set type="module"
npm pkg set main="index.mjs"
npm pkg set scripts.start="node index.mjs"
```

3. Cloud SQL Node.js コネクタの依存関係をインストールします。

```bash
npm install @google-cloud/cloud-sql-connector
```

4. PostgreSQL データベースと対話するために、`pg` をインストールします。

```bash
npm install pg
```

5. HTTP リクエストを受信するために、 `express` をインストールします。

```bash
npm install express
```

6. アプリケーション コードを含む `index.mjs` ファイルを確認します。 このコードでは次のことが可能です:

- HTTPリクエストを受け入れる
- データベースに接続する
- HTTPリクエストの時刻をデータベースに保存する
- 最後の 5 つのリクエストの時間を返す

Cloud Shell で次のコマンドを実行します:

```bash
cat ./index.mjs
```

[**index.mjs**](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/index.mjs)

```javascript
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

```

このコードは、PORT 環境変数で定義されたポートをリッスンする基本的な Web サーバーを作成します。 これで、アプリケーションをデプロイする準備が整いました。

## 6. **Cloud Run アプリケーションのデプロイ**

以下のコマンドを実行して、アプリケーションを Cloud Run にデプロイします。コマンドのオプションはそれぞれ以下の意味を持ちます:

- **–region**: インスタンスの地域の場所 (例: us-central1、asia-east1、us-east1)。
- **–source**: デプロイするソースコードのパス。この場合、現在のフォルダー helloworld 内のソースコードを参照します。
- **–set-env-vars**: アプリケーションを Cloud SQL データベースに誘導するためにアプリケーションで使用される環境変数を設定します。
- **–service-account**: このラボの冒頭で作成した Cloud SQL データベースに接続する権限を持つサービス アカウントに Cloud Run サービスを関連付けます。
- **–allow-unauthenticated**: アプリケーションがインターネットからアクセスできるように、未認証のリクエストを許可します。

```bash
gcloud run deploy helloworld \
  --region=asia-northeast1 \
  --source=. \
  --set-env-vars INSTANCE_CONNECTION_NAME="${GOOGLE_CLOUD_PROJECT}:asia-northeast1:quickstart-instance" \
  --set-env-vars DB_NAME="quickstart_db" \
  --set-env-vars DB_USER="quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam" \
  --service-account="quickstart-service-account@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --allow-unauthenticated
```

プロンプトが表示されたら、「**y**」と「**Enter**」を押して続行することを確認します:

```bash
Do you want to continue (Y/n)? y
```

数分後、アプリケーションにアクセスするための URL が出力されます

URL に移動して、アプリケーションの動作を確認します。 URL にアクセスするか、ページを更新するたびに、最近 5 件の訪問が JSON 形式で返されることがわかります。

## 7. **おめでとうございます!**

Cloud SQL 上で実行されている PostgreSQL データベースに接続できる Node.js アプリケーションを Cloud Run にデプロイしました。

### **このセクションでカバーした内容**
- Cloud SQL for PostgreSQL データベースの作成
- Cloud Run への Node.js アプリケーションのデプロイ
- Cloud SQL Node.js コネクタを使用したアプリケーションの Cloud SQL への接続

### **クリーンアップ**
このチュートリアルで使用するリソースに対して料金が発生しないようにするには、リソースを含むプロジェクトを削除するか、プロジェクトを保持して個々のリソースを削除します。プロジェクト全体を削除したい場合は、次のコマンドを実行できます。なお、このラボは後半に続きます。後半に進む方はここではプロジェクトの削除は行わず、後半のラボ終了時に削除頂いても構いません。

```bash
gcloud projects delete ${GOOGLE_CLOUD_PROJECT}
```

おつかれさまでした。

---

## Cloud Run からフルマネージドデータベース - Cloud Spanner & Cloud Firestore につなげよう

### 1. 概要
このセクションでは、サーバーレスデータベース (Spanner と Firestore) を Cloud Run で稼働しているアプリケーション (Go と Node.js) とつなげます。Cymbal Eats アプリケーションには、Cloud Run で実行されている複数のサービスが含まれています。
このハンズオンでは、[Cloud Spanner](https://cloud.google.com/spanner) (リレーショナル データベース) と [Cloud Firestore](https://cloud.google.com/firestore) ( NoSQL ドキュメント データベース) を使用するようにサービスを構成します。 データ層とアプリケーション ランタイムにサーバーレス製品を利用すると、すべてのインフラストラクチャ管理を抽象化し、オーバーヘッドを気にせずにアプリケーションの構築に集中できます。

### 2. このハンズオンで学べること
このハンズオンでは以下について学習します:

- Cloud Spanner
  - Cloud Spanner マネージドサービス を有効にする
  - アプリをデプロイして Spanner に接続する
- Cloud Firestore
  - Cloud Firestore マネージドサービスを有効にする
  - アプリをデプロイして Firestore に接続する


## 3. セットアップと要件

### Google Cloud Project の準備

ハンズオンを行う Google Cloud プロジェクトをまだ作成されていない場合は、[こちらのリンク](https://console.cloud.google.com/projectcreate) から新しいプロジェクトを作成してください。

**なるべく新しいプロジェクトが望ましいです。**

### 環境の準備

1. プロジェクト ID 変数の定義

```bash
export PROJECT_ID=$(gcloud config get-value project)
export PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
export SPANNER_INSTANCE=inventory-instance
export SPANNER_DB=inventory-db
export REGION=asia-northeast1
export SPANNER_CONNECTION_STRING=projects/$PROJECT_ID/instances/$SPANNER_INSTANCE/databases/$SPANNER_DB
```

2. 各 API の有効化 - Spanner, Cloud Run, Cloud Build, Artifact Registry

```bash
gcloud services enable \
     compute.googleapis.com \
     spanner.googleapis.com \
     run.googleapis.com \
     cloudbuild.googleapis.com \
     artifactregistry.googleapis.com \
     firestore.googleapis.com \
     appengine.googleapis.com \
     artifactregistry.googleapis.com
```

3. リポジトリのクローン

```bash
git clone https://github.com/GoogleCloudPlatform/cymbal-eats.git
```

4. ディレクトリ移動

```bash
cd cymbal-eats/inventory-service/spanner
```


## 4. Cloud Spanner インスタンスの作成と設定

Spanner は、インベントリ サービスのバックエンド リレーショナル データベースです。 次の手順で、Spanner インスタンス、データベース、およびスキーマを作成します。

### インスタンスの作成

1. Spanner インスタンスを作成

```bash
gcloud spanner instances create $SPANNER_INSTANCE --config=regional-${REGION} \
--description="Cymbal Menu Inventory" --nodes=1
```
Example Output
```
Creating instance...done.
```

2. Spanner インスタンスが正しく設定されているか確認

```bash
gcloud spanner instances list
```

Example output
```
NAME: inventory-instance
DISPLAY_NAME: Cymbal Menu Inventory
CONFIG: regional-us-east1
NODE_COUNT: 1
PROCESSING_UNITS: 100
STATE: READY
```

### データベースとスキーマの作成

新しいデータベースを作成し、[Google 標準 SQL のデータ定義言語 (DDL)](https://cloud.google.com/spanner/docs/reference/standard-sql/data-defining- language) を使用してデータベース スキーマを作成します。

1. DDL file を作成

```bash
echo "CREATE TABLE InventoryHistory (ItemRowID STRING (36) NOT NULL, ItemID INT64 NOT NULL, InventoryChange INT64, Timestamp TIMESTAMP) PRIMARY KEY(ItemRowID)" >> table.ddl
```

2. Spanner database を作成

```bash
gcloud spanner databases create $SPANNER_DB \
--instance=$SPANNER_INSTANCE \
--ddl-file=table.ddl
```

Example output

```
Creating database...done.
```

### データベースの状態とスキーマを確認する

1. データベースの状態を表示する

```bash
gcloud spanner databases describe $SPANNER_DB \
--instance=$SPANNER_INSTANCE
```

Example output

```
createTime: '2022-04-22T15:11:33.559300Z'
databaseDialect: GOOGLE_STANDARD_SQL
earliestVersionTime: '2022-04-22T15:11:33.559300Z'
encryptionInfo:
- encryptionType: GOOGLE_DEFAULT_ENCRYPTION
name: projects/cymbal-eats-7-348013/instances/menu-inventory/databases/menu-inventory
state: READY
versionRetentionPeriod: 1h
```

> Note: データベースの状態が READY と表示される

2. データベースのスキーマを表示する

```bash
gcloud spanner databases ddl describe $SPANNER_DB \
--instance=$SPANNER_INSTANCE
```

Example output

```
CREATE TABLE InventoryHistory (
  ItemRowID STRING(36) NOT NULL,
  ItemID INT64 NOT NULL,
  InventoryChange INT64,
  TimeStamp TIMESTAMP,
) PRIMARY KEY(ItemRowID);
```

> **Note**: データベースには 4 つの列があります。 ItemRowID が主キーです。
> [Spanner 概要コンソール](https://console.cloud.google.com/spanner/instances/inventory-instance/details/databases) で詳細を確認することもできます。


## 5. Spanner インテグレーション

このセクションでは、Spanner をアプリケーションに統合する方法を学習します。 さらに、SQL Spanner は [クライアント ライブラリ](https://cloud.google.com/spanner/docs/reference/libraries)、[JDBC ドライバー](https://cloud.google.com/spanner/docs/jdbc-drivers)、[R2DBC ドライバー](https://cloud.google.com/spanner/docs/use-oss-r2dbc)、[REST API](https://cloud.google.com/spanner/docs/reference/rest) と [RPC API](https://cloud.google.com/spanner/docs/reference/rpc) を提供しており、Spanner を任意のアプリケーションに統合できます。

次のセクションでは、Go クライアント ライブラリを使用して、Spanner でデータをインストール、認証、および変更します。

### クライアント ライブラリのインストール

[Cloud Spanner クライアント ライブラリ](https://cloud.google.com/spanner/docs/reference/libraries#create-service-account-gcloud) では、サービス アカウントの資格情報を見つけるのにアプリケーションのデフォルト認証情報 (ADC) を自動的に使用しており、Cloud Spanner との統合が容易になります。

> Note: コードを更新すると、スターター コードにエラーが発生します。 これらのエラーは無視してかまいません。

### 認証のセットアップ

Google Cloud CLI と Google Cloud クライアント ライブラリは、Google Cloud 上で実行されていることを自動的に検出し、現在の Cloud Run リビジョンのランタイム サービス アカウントを使用します。 この戦略はアプリケーションのデフォルト資格情報と呼ばれ、複数の環境間でのコードの移植性を可能にします。

しかしながら、デフォルトのサービス アカウントではなく、ユーザー管理のサービス アカウントを割り当てて、専用の ID を作成することがベストです。

1. Spanner データベース管理者ロールをサービス アカウントに付与します。

```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
--member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
--role="roles/spanner.databaseAdmin"
```

Example output
```
Updated IAM policy for project [cymbal-eats-6422-3462].
[...]
```

> Spanner データベース管理者ロールを使用すると、サービス アカウントは次のことができます。
> - プロジェクト内のすべての Cloud Spanner インスタンスを取得/リストします。
> - インスタンス内のデータベースを作成/リスト/ドロップします。
> - プロジェクト内のデータベースへのアクセスを許可/取り消します。
> - プロジェクト内のすべての Cloud Spanner データベースの読み取りと書き込み。

### クライアントライブラリの使用

Spanner クライアント ライブラリは、Spanner との統合の複雑さを抽象化し、多くの一般的なプログラミング言語で利用できます。

#### Spanner クライアントを作成

Spanner クライアントは、Cloud Spanner データベースに対してデータを読み書きするためのクライアントです。 クライアントは、Close メソッドを除き、同時に使用しても安全です。

以下のスニペットは Spanner クライアントの作成です

**[main.go](https://github.com/GoogleCloudPlatform/cymbal-eats/blob/main/inventory-service/spanner/main.go#L47-L61)**

```golang
var dataClient *spanner.Client
...
dataClient, err = spanner.NewClient(ctx, databaseName)
```

クライアントはデータベースとのコネクションと考えることができ、Cloud Spanner とのやり取りはすべてクライアントを経由する必要があります。 通常、アプリケーションの起動時にクライアントを作成し、そのクライアントを再利用してトランザクションの読み取り、書き込み、実行を行います。 各クライアントは Cloud Spanner のリソースを使用します。

### データの変更

Spanner データベースのデータを挿入、更新、削除するには、複数の方法があります。 利用可能な方法を以下に示します。

- [Google Cloud Console](https://cloud.google.com/spanner/docs/modify-data)
- [gcloud CLI](https://cloud.google.com/spanner/docs/modify-gcloud)
- [DML](https://cloud.google.com/spanner/docs/modify-gcloud#modifying_data_using_dml)
- [Mutations](https://cloud.google.com/spanner/docs/modify-mutation-api)

このハンズオンでは、ミューテーションを使用してデータを変更します

### Spanner のミューテーション

[Mutation](https://pkg.go.dev/cloud.google.com/go/spanner/#Mutation) は、ミューテーション操作用のコンテナです。 ミューテーションは、Cloud Spanner が Cloud Spanner データベース内のさまざまな行やテーブルにアトミックに適用する一連の挿入、更新、削除を表します。

**[main.go](https://github.com/GoogleCloudPlatform/cymbal-eats/blob/main/inventory-service/spanner/main.go#L148-L153)**

```golang
m := []*spanner.Mutation{}

m = append(m, spanner.Insert(
        "inventoryHistory",
         inventoryHistoryColumns,
        []interface{}{uuid.New().String(), element.ItemID, element.InventoryChange, time.Now()}))
```

このコードスニペットは、在庫履歴テーブルに新しい行を挿入しています。

### デプロイとテスト

Spanner が構成され、主要なコード要素を確認してました。プリケーションを Cloud Run にデプロイしましょう。

### アプリケーションを Cloud Run にデプロイする

Cloud Run では、1 つのコマンドでコードを自動的にビルド、プッシュ、デプロイできます。 次のコマンドでは、「run」サービスで「deploy」コマンドを呼び出し、前に作成した SPANNER_CONNECTION_STRING など、実行中のアプリケーションで使用される変数を渡します。

1. 「ターミナルを開く」をクリックします
2. インベントリ サービスを Cloud Run にデプロイする

```bash
gcloud run deploy inventory-service \
    --source . \
    --region $REGION \
    --update-env-vars SPANNER_CONNECTION_STRING=$SPANNER_CONNECTION_STRING \
    --allow-unauthenticated \
    --project=$PROJECT_ID \
    --quiet
```

Example output

```
Service [inventory-service] revision [inventory-service-00001-sug] has been deployed and is serving 100 percent of traffic.
Service URL: https://inventory-service-ilwytgcbca-uk.a.run.app
```

> **Note**: 続行するように求められたら、「Y」と入力します

3. サービス URL を保存する

```bash
INVENTORY_SERVICE_URL=$(gcloud run services describe inventory-service \
  --platform managed \
  --region $REGION \
  --format=json | jq \
  --raw-output ".status.url")
```

### Cloud Run アプリケーションをテストする

#### アイテムの挿入

Cloud Shell で次のコマンドを入力します。

```bash
POST_URL=$INVENTORY_SERVICE_URL/updateInventoryItem
curl -i -X POST ${POST_URL} \
--header 'Content-Type: application/json' \
--data-raw '[
    {
        "itemID": 1,
        "inventoryChange": 5
    }
]'
```

Example output

```
HTTP/2 200
access-control-allow-origin: *
content-type: application/json
x-cloud-trace-context: 10c32f0863d26521497dc26e86419f13;o=1
date: Fri, 22 Apr 2022 21:41:38 GMT
server: Google Frontend
content-length: 2

OK
```

### アイテムをクエリする

1. インベントリサービスをクエリする

```
GET_URL=$INVENTORY_SERVICE_URL/getAvailableInventory
curl -i ${GET_URL}
```

Example response

```
HTTP/2 200
access-control-allow-origin: *
content-type: text/plain; charset=utf-8
x-cloud-trace-context: b94f921e4c2ae90210472c88eb05ace8;o=1
date: Fri, 22 Apr 2022 21:45:50 GMT
server: Google Frontend
content-length: 166

[{"ItemID":1,"Inventory":5}]
```

## 6. Spanner コンセプト

Cloud Spanner は、宣言型 SQL ステートメントを使用してデータベースにクエリを実行します。 SQL ステートメントは、結果がどのように得られるかについては説明せずに、ユーザーが望むものを示します。

1. ターミナルで次のコマンドを入力して、以前に作成したレコードをテーブルにクエリします。

```bash
gcloud spanner databases execute-sql $SPANNER_DB \
--instance=$SPANNER_INSTANCE \
--sql='SELECT * FROM InventoryHistory WHERE ItemID=1'
```

Example output

```
ItemRowID: 1
ItemID: 1
InventoryChange: 3
Timestamp:
```

### クエリ実行プラン

[クエリ実行プラン](https://cloud.google.com/spanner/docs/query-execution-plans) は、Spanner が結果を取得するために使用する一連のステップです。 特定の SQL ステートメントの結果を取得するには、いくつかの方法があります。 クエリ実行プランには、コンソールとクライアント ライブラリからアクセスできます。 Spanner が SQL クエリをどのように処理するかを確認するには、次の手順を実行します:

1. コンソールで、Cloud Spanner インスタンス ページを開きます。
2. Cloud Spanner インスタンスに移動します
3. Cloud Spanner インスタンスの名前をクリックします。 データベース セクションから、クエリを実行するデータベースを選択します。
4. 「クエリ」をクリックします。
5. クエリエディタに次のクエリを入力します。

```sql
SELECT * FROM InventoryHistory WHERE ItemID=1
```

6. 「実行」をクリックします。
7. 「説明」をクリックします。

Cloud Console には、クエリの実行プランが視覚的に表示されます。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-spanner-concept.png)

> 概念的には、実行計画は関係演算子のツリーです。 各演算子は入力から行を読み取り、出力行を生成します。 実行のルートが SQL クエリの結果として返されます。

### クエリオプティマイザー

Cloud Spanner クエリ オプティマイザーは、実行プランを比較し、最も効率的な実行プランを選択します。 時間の経過とともに、クエリ オプティマイザーは進化し、クエリ実行計画の選択肢が広がり、それらの選択肢を知らせる推定の精度が向上し、より効率的なクエリ実行計画につながります。

Cloud Spanner は、オプティマイザーの更新を新しいクエリ オプティマイザー バージョンとしてロールアウトします。 デフォルトでは、各データベースはオプティマイザーの最新バージョンがリリースされてから 30 日以内にそのバージョンの使用を開始します。

gcloud Spanner でクエリを実行するときに使用されるバージョンを確認するには、–query-mode フラグを PROFILE に設定します。

1. 次のコマンドを入力して、オプティマイザーのバージョンを表示します。

```bash
gcloud spanner databases execute-sql $SPANNER_DB --instance=$SPANNER_INSTANCE \
--query-mode=PROFILE --sql='SELECT * FROM InventoryHistory'
```

Example output

```
TOTAL_ELAPSED_TIME: 6.18 msecs
CPU_TIME: 5.17 msecs
ROWS_RETURNED: 1
ROWS_SCANNED: 1
OPTIMIZER_VERSION: 3
 RELATIONAL Distributed Union
 (1 execution, 0.11 msecs total latency)
 subquery_cluster_node: 1
    |
    +- RELATIONAL Distributed Union
    |  (1 execution, 0.09 msecs total latency)
    |  call_type: Local, subquery_cluster_node: 2
    |   |
    |   \- RELATIONAL Serialize Result
    |      (1 execution, 0.08 msecs total latency)
    |       |
    |       +- RELATIONAL Scan
    |       |  (1 execution, 0.08 msecs total latency)
    |       |  Full scan: true, scan_target: InventoryHistory, scan_type: TableScan
    |       |   |
    |       |   +- SCALAR Reference
    |       |   |  ItemRowID
    |       |   |
    |       |   +- SCALAR Reference
    |       |   |  ItemID
    |       |   |
    |       |   +- SCALAR Reference
    |       |   |  InventoryChange
    |       |   |
    |       |   \- SCALAR Reference
    |       |      Timestamp
    |       |
    |       +- SCALAR Reference
    |       |  $ItemRowID
    |       |
    |       +- SCALAR Reference
    |       |  $ItemID
    |       |
    |       +- SCALAR Reference
    |       |  $InventoryChange
    |       |
    |       \- SCALAR Reference
    |          $Timestamp
    |
    \- SCALAR Constant
       true

ItemRowID: 1
ItemID: 1
InventoryChange: 3
Timestamp:
```

> 現在のバージョンはバージョン 3 に設定されています。最新バージョンを見つけるには、[バージョン履歴](https://cloud.google.com/spanner/docs/query-optimizer/overview#version-history) を確認してください。

#### オプティマイザーのバージョンを更新する

このラボの時点での最新バージョンはバージョン 4 です。次に、クエリ オプティマイザーにバージョン 4 を使用するように Spanner テーブルを更新します。

2. オプティマイザを更新する

```bash
gcloud spanner databases ddl update $SPANNER_DB \
--instance=$SPANNER_INSTANCE \
--ddl='ALTER DATABASE InventoryHistory
SET OPTIONS (optimizer_version = 4)'
```

Example ouput

```
Schema updating...done.
```

3. 次のコマンドを入力して、オプティマイザーのバージョン更新を表示します。

```bash
gcloud spanner databases execute-sql $SPANNER_DB --instance=$SPANNER_INSTANCE \
--query-mode=PROFILE --sql='SELECT * FROM InventoryHistory'
```

Example output

```
TOTAL_ELAPSED_TIME: 8.57 msecs
CPU_TIME: 8.54 msecs
ROWS_RETURNED: 1
ROWS_SCANNED: 1
OPTIMIZER_VERSION: 4
[...]
```

> `OPTIMIZER_VERSION` がバージョン 4 に更新されました

#### Metrics Explorer でクエリ オプティマイザーのバージョンを視覚化する

Cloud コンソール の Metrics Explorer を使用して、データベース インスタンスの **クエリ数** を視覚化できます。 各データベースでどのオプティマイザのバージョンが使用されているかを確認できます。

1. Cloud コンソール のモニタリングに移動し、左側のメニューで [Metrics Explorer](https://cloud.google.com/monitoring/charts/metrics-explorer#find-me) を選択します。

2. [**リソース タイプ**] フィールドで、[Cloud Spanner インスタンス] を選択します。

3. [**メトリック**] フィールドで、[クエリ数] を選択して [適用] を選択します。

4. [**グループ化**] フィールドで、データベース、optimizer_version、ステータスを選択します。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-metrics-explorer.png)


## 7. Firestore の作成と構成

Firestore は、自動スケーリング、高パフォーマンス、アプリケーション開発の容易さのために構築された NoSQL ドキュメント データベースです。 Firestore インターフェースには従来のデータベースと同じ機能が多くありますが、NoSQL データベースはデータ オブジェクト間の関係を記述する点で異なります。

以降のタスクでは、Firestore を利用した、注文サービス Cloud Run アプリケーションを作成する手順を説明します。 注文サービスは、注文を開始する前に、前のセクションで作成した在庫サービスを呼び出して、Spanner データベースにクエリを実行します。 このサービスにより、十分な在庫が存在し、注文に対応できることが保証されます。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-firestore.png)


## 8. Firestore コンセプト

### データモデル

Firestore データベースは、コレクションとドキュメントで構成されます。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-firestore02.png)

#### ドキュメント

各ドキュメントには、一連のキーと値のペアが含まれています。 Firestore は、小さなドキュメントの大規模なコレクションを保存するために最適化されています。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-firestore03.png)

> 上の例では、注文 ID ドキュメントに 4 つのキーと値のペアが含まれています。 キー orderItems には、キーと値のペアの配列が含まれます。

#### コレクション

すべてのドキュメントをコレクションに保存する必要があります。 ドキュメントには、文字列のようなプリミティブなフィールドやリストのような複雑なオブジェクトを含む、サブコレクションとネストされたオブジェクトを含めることができます。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-firestore04.png)

> 上記の例では、注文 ID ドキュメントが注文コレクションに保存されています。

### Firestore データベースの作成

1. Firestore データベースを作成する

```bash
gcloud firestore databases create --location=$REGION
```

Example ouput

```
Success! Selected Google Cloud Firestore Native database for cymbal-eats-6422-3462
```

> 作成した新しい Firestore データベースは現在空です。 新しいデータベースには、誰でも読み取り操作を実行でき、データベースへの書き込みを禁止するデフォルトのセキュリティ ルール セットもあります。

## 9. Firestore をアプリケーションに統合する

このセクションでは、サービス アカウントの更新、Firestore アクセス サービス アカウントの追加、Firestore セキュリティ ルールの確認と展開、Firestore でのデータの変更方法の確認を行います。

### 認証を設定する

1. データストア ユーザー ロールをサービス アカウントに付与する

```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/datastore.user"
```

Example output

```
Updated IAM policy for project [cymbal-eats-6422-3462].
```

> Datastore ユーザー ロールは、Firestore データベース内のデータへの読み取り/書き込みアクセスを許可します。

#### Firestore セキュリティルール

セキュリティ ルールは、アクセス制御とデータ検証を表現力豊かでわかりやすい形式で提供します。

1. order-service/starter-code ディレクトリに移動します。

```bash
cd ~/cymbal-eats/order-service
```

2. クラウドエディターで `firestore.rules` ファイルを開きます。

```bash
cat firestore.rules
```

**[firestore.rules](https://github.com/GoogleCloudPlatform/cymbal-eats/blob/main/order-service/firestore.rules)**

```
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents { ⇐ All database
    match /{document=**} { ⇐ All documents
      allow read: if true; ⇐ Allow reads
    }
    match /{document=**} {
      allow write: if false; ⇐ Deny writes
    }
  }
}
```

> firestore.rules ファイルには、Firestore データベース内のすべてのテーブルに対するすべての読み取り操作を許可し、すべての書き込み操作を拒否するルールが含まれています。 Firestore ルールの詳細については、[セキュリティ ルールの概要](https://cloud.google.com/firestore/docs/security/get-started) を参照してください。

**警告**: Firestore ストレージへのアクセスを制限することがベスト プラクティスです。 このラボでは、すべての読み取りが許可されています。 これは推奨される運用構成ではありません。

### Firestore マネージド サービスを有効にする

1. 「ターミナルを開く」をクリックします

2. 現在のプロジェクト ID を使用して .firebaserc ファイルを作成します。 デプロイ ターゲットの設定は、プロジェクト ディレクトリの .firebaserc ファイルに保存されます。

**[firebaserc.tmpl](https://github.com/GoogleCloudPlatform/cymbal-eats/blob/main/order-service/firebaserc.tmpl)**

```bash
sed "s/PROJECT_ID/$PROJECT_ID/g" firebaserc.tmpl > .firebaserc
```

2. Firebase バイナリをダウンロードする

```bash
curl -sL https://firebase.tools | upgrade=true bash
```

Example output

```
-- Checking for existing firebase-tools on PATH...
Your machine already has firebase-tools@10.7.0 installed. Nothing to do.
-- All done!
```

3. Firestore ルールをデプロイする

```bash
firebase deploy
```

Example Output

```
=== Deploying to 'cymbal-eats-6422-3462'...

i  deploying firestore
i  cloud.firestore: checking firestore.rules for compilation errors...
✔  cloud.firestore: rules file firestore.rules compiled successfully
i  firestore: uploading rules firestore.rules...
✔  firestore: released rules firestore.rules to cloud.firestore

✔  Deploy complete!

Project Console: https://console.firebase.google.com/project/cymbal-eats-6422-3462/overview
```

> Cloud Firestore セキュリティ ルールの更新が新しいクエリとリスナーに影響を与えるまでに最大 1 分かかる場合があります。 ただし、変更が完全に反映され、アクティブなリスナーに影響を与えるまでに最大 10 分かかる場合があります。

### データを変更する

コレクションとドキュメントは Firestore で暗黙的に作成されます。 コレクション内のドキュメントにデータを割り当てるだけです。 コレクションまたはドキュメントが存在しない場合は、Firestore によって作成されます。

#### Firestore にデータを追加する

Cloud Firestore にデータを書き込む方法はいくつかあります:

- 文書識別子を明示的に指定して、コレクション内の文書のデータを設定します
- 新しいドキュメントをコレクションに追加します。 この場合、Cloud Firestore はドキュメント識別子を自動的に生成します
- 自動生成された識別子を使用して空のドキュメントを作成し、後でそれにデータを割り当てます

次のセクションでは、set メソッドを使用してドキュメントを作成する方法を説明します。

#### ドキュメントをセットする

`set()` メソッドを使用してドキュメントを作成します。 `set()` メソッドでは、作成するドキュメントの ID を指定する必要があります。

以下のコードスニペットを見てください。

**[index.js](https://github.com/GoogleCloudPlatform/cymbal-eats/tree/main/order-service/index.js#L89-L102)**

```javascript
const orderDoc = db.doc(`orders/123`);
await orderDoc.set({
    orderNumber: 123,
    name: Anne,
    address: 555 Bright Street,
    city: Mountain View,
    state: CA,
    zip: 94043,
    orderItems: [id: 1],
    status: 'New'
  });
```

このコードは、ユーザーが生成したドキュメント ID 123 を指定してドキュメントを作成します。ユーザーに代わって Firestore に ID を生成させるには、`add()` または `create()` メソッドを使用します。

> `set()` を使用すると、ドキュメントが存在しない場合にドキュメントが作成されます。 ドキュメントが存在する場合、その内容は新しく提供されたデータで上書きされます。

#### ドキュメントを更新する

更新メソッド `update()` を使用すると、ドキュメント全体を上書きせずに、一部のドキュメント フィールドを更新できます。

以下のスニペットでは、コードは順序 123 を更新します。

**[index.js](https://github.com/GoogleCloudPlatform/cymbal-eats/tree/main/order-service/index.js#L62-L63)**

```javascript
const orderDoc = db.doc(`orders/123`);
await orderDoc.update(name: "Anna");
```

#### ドキュメントを削除する

Firestore では、ドキュメントからコレクション、ドキュメント、または特定のフィールドを削除できます。 ドキュメントを削除するには、`delete()` メソッドを使用します。

以下のスニペットは注文 123 を削除します。

**[index.js](https://github.com/GoogleCloudPlatform/cymbal-eats/tree/main/order-service/index.js#L50-L51)**

```javascript
const orderDoc = db.doc(`orders/123`);
await orderDoc.delete();
```

> **注意**: ドキュメントを削除しても、そのサブコレクションは削除されません。

## 10. デプロイとテスト

このセクションでは、アプリケーションを Cloud Run にデプロイし、作成、更新、削除メソッドをテストします。

### アプリケーションを Cloud Run にデプロイする

1. URL を変数 INVENTORY_SERVICE_URL に保存して、Inventory Service と統合します

```bash
INVENTORY_SERVICE_URL=$(gcloud run services describe inventory-service \
 --region=$REGION \
 --format=json | jq \
 --raw-output ".status.url")
```

> 注文サービスは在庫サービスと通信して、在庫が存在し、注文が履行できることを確認する必要があります。 このステップでは、在庫サービスの URL を変数に保存し、この変数が環境変数として注文サービスに渡されます。

2. 注文サービスをデプロイする

```bash
gcloud run deploy order-service \
  --source . \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --project=$PROJECT_ID \
  --set-env-vars=INVENTORY_SERVICE_URL=$INVENTORY_SERVICE_URL \
  --quiet
```

Example output

```
[...]
Done.
Service [order-service] revision [order-service-00001-qot] has been deployed and is serving 100 percent of traffic.
Service URL: https://order-service-3jbm3exegq-uk.a.run.app
```

### Cloud Run アプリケーションをテストする

#### ドキュメントを作成する

1. 注文サービス アプリケーションの URL をテスト用の変数に保存します

```bash
ORDER_SERVICE_URL=$(gcloud run services describe order-service \
  --platform managed \
  --region $REGION \
  --format=json | jq \
  --raw-output ".status.url")
```

2. 注文リクエストを作成し、新しい注文を Firestore データベースにポストします

```bash
curl --request POST $ORDER_SERVICE_URL/order \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "Jane Doe",
         "email": "Jane.Doe-cymbaleats@gmail.com",
    "address": "123 Maple",
    "city": "Buffalo",
    "state": "NY",
    "zip": "12346",
    "orderItems": [
        {
            "id": 1
        }
    ]
}'
```

Example output

```
{"orderNumber":46429}
```

#### 後で使用できるように注文番号を保存します

```bash
export ORDER_NUMBER=<value_from_output>
```

#### 結果を見る

Firestore で結果を表示する

1. [Firestore コンソール](https://console.cloud.google.com/firestore) に移動します。

2. 「データ」をクリックします。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-firestore05.png)

### ドキュメントの更新

送信された注文には数量が含まれていませんでした。

1. レコードを更新し、数量のキーと値のペアを追加します。

```bash
curl --location -g --request PATCH $ORDER_SERVICE_URL/order/${ORDER_NUMBER} \
--header 'Content-Type: application/json' \
--data-raw '{
"orderItems": [
        {
            "id": 1,
            "quantity": 1
        }
    ]
}'
```

Example output

```
{"status":"success"}
```

#### 結果を見る

Firestore で結果を表示する

1. [Firestore コンソール](https://console.cloud.google.com/firestore) に移動します。

2. 「データ」をクリックします。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-firestore06.png)

> When updating the NoSQL structure in Firestore using patch(), only the items which are passed in the call are updated.

### Delete a document

1. Delete item 46429 from the Firestore orders collection

```bash
curl --location -g --request DELETE $ORDER_SERVICE_URL/order/${ORDER_NUMBER}
```

#### 結果を見る

Firestore で結果を表示する

1. [Firestore コンソール](https://console.cloud.google.com/firestore) に移動します。

2. 「データ」をクリックします。

![](https://github.com/google-cloud-japan/gig-training-materials/blob/main/gig07-02/img/gig07_02-firestore07.png)

> ドキュメント 46429 は削除されましたが、注文のコレクションは残ります。

## 11. おめでとうございます!

おめでとうございます、こちらでハンズオンは終了です!

次のステップ:
Cymbal Eats コードラボは他にもありますのでぜひトライしてください。:

- [Eventarc を使用したクラウド ワークフローのトリガー](https://codelabs.developers.google.com/eventarc-workflows-cloud-run)
- [クラウド ストレージからのイベント処理のトリガー](https://codelabs.developers.google.com/triggering-cloud-functions-from-cloud-storage)
- [Cloud Run からプライベート CloudSQL への接続](https://codelabs.developers.google.com/connecting-to-private-cloudsql-from-cloud-run)
- [Identity Aware Proxy (IAP) を使用した安全なサーバーレス アプリケーション](https://codelabs.developers.google.com/secure-serverless-application-with-identity-aware-proxy)
- [Cloud Scheduler を使用した Cloud Run ジョブのトリガー](https://codelabs.developers.google.com/cloud-run-jobs-and-cloud-scheduler)
- [Cloud Run への安全なデプロイ](https://codelabs.developers.google.com/secure-cloud-run-deployment)
- [Cloud Run Ingress トラフィックの保護](https://codelabs.developers.google.com/cloud-run-ingress-deployment)
- [GKE Autopilot からプライベート AlloyDB への接続](https://codelabs.developers.google.com/connecting-to-private-alloydb-from-gke-autopilot)

### クリーンアップ

このチュートリアルで使用するリソースに対して Google Cloud アカウントに料金が発生しないようにするには、リソースを含むプロジェクトを削除するか、プロジェクトを保持して個々のリソースを削除します。

#### プロジェクトの削除

このチュートリアルで使用するリソースに対して不要な料金が発生しないようにするには、リソースを含むプロジェクトを削除するか、プロジェクトを保持して個々のリソースを削除します。プロジェクト全体を削除したい場合は、次のコマンドを実行できます:

```bash
gcloud projects delete ${GOOGLE_CLOUD_PROJECT}
```

おつかれさまでした。
