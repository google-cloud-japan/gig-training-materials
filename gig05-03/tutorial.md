<!-- TODO: Add the instruction for setting up the environmental variables? -->

# [GIG ハンズオン] **Cloud Code、Cloud Build、Google Cloud Deploy、GKE を使用したアプリの開発と配信**

[オリジナル公式ドキュメント](https://cloud.google.com/architecture/app-development-and-delivery-with-cloud-code-gcb-cd-and-gke?hl=ja)

## 目次 - Table of Contents
- [解説] ハンズオンの内容と目的
- アーキテクチャの概要
- 目標
- 費用
- Google Cloud プロジェクトの選択
- 環境を準備する
  - 権限を設定する
  - GKEクラスタを作成する
  - IDEを開いてリポジトリのクローンを作成する
  - ソースコード用のリポジトリとコンテナ用のリポジトリを作成する
- CI / CD パイプラインを構成する
- デベロッパー ワークスペース内でアプリケーションを変更する
  - アプリケーションをビルドし、テストして、実行する
  - 変更する
  - コードを commit する
- 本番環境に変更をデプロイする
  - CI / CD パイプラインを開始してステージング環境にデプロイする
  - リリースを本番環境に昇格させる
- クリーンアップ
  - オプション1: プロジェクトを削除する
  - オプション2: 個々のリソースを削除する
- 次のステップ

## [解説] **ハンズオンの内容と目的**

このチュートリアルでは、Google Cloud ツールの統合セットを使用して、開発、継続的インテグレーション(CI)、継続的デリバリー(CD)システムを設定し、使用する方法について説明します。このシステムを使用すると、アプリケーションを開発し、[Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine?hl=ja)にデプロイできます。

このチュートリアルは、ソフトウェア デベロッパーとオペレーターの両方を対象としており、完了すると両方の役割を果たせます。まず、CI / CD パイプラインを設定するオペレーターの役割を果たします。このパイプラインの主要なコンポーネントは、[Cloud Build](https://cloud.google.com/build?hl=ja)、[Artifact Registry](https://cloud.google.com/artifact-registry?hl=ja)、[Google Cloud Deploy](https://cloud.google.com/deploy?hl=ja) です。

次にデベロッパーとして機能し、[Cloud Code](https://cloud.google.com/code?hl=ja) を使用してアプリケーションを変更します。デベロッパーとして機能すると、このパイプラインが提供する統合されたエクスペリエンスが表示されます。

最後に、オペレータとして機能して、アプリケーションを本番環境にデプロイする手順を実行します。

このチュートリアルは、Google Cloud での gcloud コマンドの実行と GKE へのアプリケーション コンテナのデプロイに精通していることを前提としています。

この統合システムの主な特徴は次のとおりです。

- **開発とデプロイが迅速になります。**

  デベロッパー ワークスペースで変更を検証できるため、開発ループが効率的です。自動化された CI/CD システムと環境間のパリティの向上により、本番環境に変更を展開する際に、より多くの問題を検出できるため、デプロイが高速化されます。

- **開発、ステージング、本番環境全体で一貫性の向上によるメリットを享受できます。**

  このシステムのコンポーネントは、共通の Google Cloud ツールセットを使用します。

- **さまざまな環境で構成を再利用します。**

  こうした再利用は Skaffold を使用して行われますが、Skaffold は異なる環境に対して共通の構成形式を認めます。また、デベロッパーとオペレーターが同じ構成を更新して使用することもできます。

- **ワークフローの早い段階でガバナンスを適用します。**

  このシステムでは、本番環境、CI システム、開発環境でガバナンスに関する検証テストが適用されます。開発環境では、ガバナンスを適用することで、問題を検出し、早期に修正できます。

- **独自のツールでソフトウェア配信を管理します。**

  継続的デリバリーはフルマネージドであり、CD パイプラインのステージをレンダリングやデプロイの細部から分離します。

## **アーキテクチャの概要**

次の図に、このチュートリアルで使用するリソースを示します。

![](https://raw.githubusercontent.com/google-cloud-japan/gig-training-materials/16-gig-5-3-%E3%82%B3%E3%83%B3%E3%83%86%E3%83%B3%E3%83%84%E4%BD%9C%E6%88%90/gig05-03/img/app-development-and-delivery-with-cloud-code-gcb-cd-and-gke.png)

このパイプラインを構成する 3 つの主要コンポーネントは次のとおりです。

1. 開発ワークスペースとしての **Cloud Code**

    このワークスペースの一部として、[minikube](https://minikube.sigs.k8s.io/docs/) で実行される開発クラスタで変更を確認できます。[Cloud Shell](https://cloud.google.com/shell?hl=ja) で Cloud Code と minikube クラスタを実行します。Cloud Shell は、ブラウザからアクセスできるオンライン開発環境です。コンピューティング リソース、メモリ、統合開発環境(IDE)を備え、Cloud Code もインストールされます。

2. アプリケーションのビルドとテストを行うための **Cloud Build**(パイプラインの「CI」部分)

    パイプラインのこの部分には、次のアクションが含まれます。

    - Cloud Build は、Cloud Build トリガーを使用して、ソース リポジトリに対する変更をモニタリングします。'
    - メインブランチに変更が commit されると、Cloud Build トリガーは次の処理を行います。
      - アプリケーション コンテナを再構築します。
      - ビルド アーティファクトを Cloud Storage バケットに配置します。
      - アプリケーション コンテナを Artifact Registry に配置します。
      - コンテナでテストを実行します。
      - Google Cloud Deploy を呼び出して、コンテナをステージング環境にデプロイします。このチュートリアルでは、ステージング環境は Google Kubernetes Engine クラスタです。
    - ビルドとテストが成功したら、Google Cloud Deploy を使用して、ステージング環境から本番環境にコンテナを昇格できます。

3. デプロイを管理するための **Google Cloud Deploy**(パイプラインの「CD」部分)

    パイプラインのこの部分では、Google Cloud Deploy が次の処理を行います。

    - [配信パイプライン](https://cloud.google.com/deploy/docs/terminology?hl=ja#delivery_pipeline)と[ターゲット](https://cloud.google.com/deploy/docs/terminology?hl=ja#target)を登録します。ターゲットはステージング クラスタと本番環境クラスタを表します。
    - Cloud Storage バケットを作成し、Skaffold レンダリング ソースとレンダリングされたマニフェストを作成したバケットに保存します。
    - ソースコードを変更するたびに(新しいリリース)(https://cloud.google.com/deploy/docs/terminology?hl=ja#release)を行います。このチュートリアルには 1 つの変更点があるため、新しいリリースを 1 つ行います。
    - アプリケーションを本番環境にデプロイします。この本番環境へのデプロイでは、運用担当者(または指定した人)がデプロイを手動で承認します。このチュートリアルでは、本番環境は Google Kubernetes Engine クラスタです。

Kubernetes ネイティブ アプリケーションの継続的な開発を容易にするコマンドライン ツールである [Skaffold](https://skaffold.dev/) は、これらのコンポーネントの基盤であり、開発、ステージング、本番環境の間で構成を共有できます。

Google Cloud はアプリケーションのソースコードを GitHub に保存します。このチュートリアルの一環として、このリポジトリのクローンを Cloud Source Repositories に作成して、CI / CD パイプラインに接続します。

このチュートリアルでは、システムのほとんどのコンポーネントで Google Cloud プロダクトを使用し、Skaffold でシステムを統合しています。Skaffold はオープンソースであるため、このような原則で Google Cloud、社内コンポーネント、サードパーティ コンポーネントを組み合わせて同様のシステムを作成できます。このソリューションはモジュール方式を採用しているため、開発パイプラインとデプロイ パイプラインの一部として段階的に導入できます。

## **目標**

オペレーターとして、次の操作を行います。

- CI パイプラインと CD パイプラインを設定します。この設定には次のものが含まれます。
  - 必要な権限を設定する。
  - ステージング環境と本番環境用の GKE クラスタを作成する。
  - ソースコード用のリポジトリを Cloud Source Repositories に作成する。
  - アプリケーション コンテナ用のリポジトリを Artifact Registry に作成する。
  - メインの GitHub リポジトリに Cloud Build トリガーを作成する。
  - Google Cloud Deploy デリバリー パイプラインとターゲットを作成する。ターゲットはステージング環境と本番環境です。
- ステージング環境にデプロイする CI / CD プロセスを開始し、本番環境に昇格させる。

開発者として、アプリケーションに変更を加えます。手順は次のとおりです。

- 事前構成された開発環境と連携するように、リポジトリのクローンを作成する。
- デベロッパー ワークスペース内でアプリケーションを変更する
- 変更をビルドおよびテストします。テストには、ガバナンスの検証テストが含まれます。
- 開発クラスタで変更を表示して検証します。このクラスタは minikube で実行されます。
- メイン リポジトリに変更を commit する。

## **費用**

このチュートリアルでは、課金対象である次の Google Cloud コンポーネントを使用します。

[Cloud Build](https://cloud.google.com/build/pricing?hl=ja)
[Google Cloud Deploy](https://cloud.google.com/deploy/pricing?hl=ja)
[Artifact Registry](https://cloud.google.com/artifact-registry/pricing?hl=ja)
[Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/pricing?hl=ja)
[Cloud Source Repositories](https://cloud.google.com/source-repositories/pricing?hl=ja)
[Cloud Storage](https://cloud.google.com/storage/pricing?hl=ja)
[料金計算ツール](https://cloud.google.com/products/calculator?hl=ja)を使うと、予想使用量に基づいて費用の見積もりを生成できます。

このチュートリアルを終了した後、作成したリソースを削除すると、それ以上の請求は発生しません。詳細については、[クリーンアップ](https://cloud.google.com/architecture/app-development-and-delivery-with-cloud-code-gcb-cd-and-gke?hl=ja#clean-up)をご覧ください。

## Google Cloud プロジェクトの選択

1. ハンズオンを行う Google Cloud プロジェクトを作成し、 Google Cloud プロジェクトを選択して **Start/開始** をクリックしてください。

**なるべく新しいプロジェクトを作成してください。**

<walkthrough-project-setup>
</walkthrough-project-setup>

2. Cloud プロジェクトに対して課金が有効になっていることを確認します。詳しくは、[プロジェクトで課金が有効になっているかどうかを確認する方法](https://cloud.google.com/billing/docs/how-to/verify-billing-enabled?hl=ja)をご覧ください。

3. Artifact Registry, Cloud Build, Google Cloud Deploy, Cloud Source Repositories, Google Kubernetes Engine, Resource Manager, and Service Networking API を有効にします。

    [API を有効にするリンク](https://console.cloud.google.com/flows/enableapi?apiid=artifactregistry.googleapis.com%2Ccloudbuild.googleapis.com%2Cclouddeploy.googleapis.com%2Csourcerepo.googleapis.com%2Ccontainer.googleapis.com%2C+cloudresourcemanager.googleapis.com%2Cservicenetworking.googleapis.com&%3Bredirect=https%3A%2F%2Fconsole.cloud.google.com&hl=ja&_ga=2.152803194.2113702237.1667787342-853816604.1666918848)

4. Google Cloud コンソールで、「Cloud Shell をアクティブにする」をクリックします。

    [Cloud Shell をアクティブにする](https://console.cloud.google.com/?cloudshell=true&hl=ja&_ga=2.144924279.2113702237.1667787342-853816604.1666918848)

    Google Cloud コンソールの下部で [Cloud Shell](https://cloud.google.com/shell/docs/how-cloud-shell-works?hl=ja) セッションが開始し、コマンドライン プロンプトが表示されます。Cloud Shell はシェル環境です。Google Cloud CLI がすでにインストールされており、現在のプロジェクトの値もすでに設定されています。セッションが初期化されるまで数秒かかることがあります。

## **環境を準備する**

このセクションでは、アプリケーション オペレーターとして、次の処理を行います。

- 必要な権限を設定する。
- ステージング環境と本番環境用の GKE クラスタを作成する。
- ソース リポジトリのクローンを作成する
- ソースコード用のリポジトリを Cloud Source Repositories に作成する。
- コンテナ アプリケーション用のリポジトリを Artifact Registry に作成します。

### **権限を設定する**

このセクションでは、CI / CD パイプラインの設定に必要な権限を付与します。

1. Cloud Shell エディタの新しいインスタンスで作業している場合は、このチュートリアルで使用するプロジェクトを指定します。

```sh
export PROJECT_ID=<walkthrough-project-id/>
gcloud config set project $PROJECT_ID
```

*PROJECT_ID* は、このチュートリアルで選択または作成したプロジェクトの ID ( <walkthrough-project-id/> )を設定します。

ダイアログが表示された場合は、[承認] をクリックします。

2. 必要なサービス アカウントを設定し、必要な権限を付与します。

  - このチュートリアルで Cloud Build と Google Cloud Deploy で使用するデフォルトの Compute Engine サービス アカウントに十分な権限が付与されていることを確認します。

    このサービス アカウントにはすでに必要な権限が付与されている場合があります。これは、デフォルトのサービス アカウントに対する自動のロール付与を無効にするプロジェクト向けのステップです。

```sh
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$(gcloud projects describe $PROJECT_ID \
    --format="value(projectNumber)")-compute@developer.gserviceaccount.com \
    --role="roles/clouddeploy.jobRunner"
```

  - Google Cloud Deploy を使用してデプロイを呼び出し、配信パイプラインとターゲットの定義を更新する Cloud Build サービス アカウント権限を付与します。

```sh
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$(gcloud projects describe $PROJECT_ID \
    --format="value(projectNumber)")@cloudbuild.gserviceaccount.com \
    --role="roles/clouddeploy.operator"
```

    この IAM ロールの詳細については、[clouddeploy.operator](https://cloud.google.com/deploy/docs/iam-roles-permissions?hl=ja#predefined_roles) ロールをご覧ください。

  - Cloud Build と Google Cloud Deploy のサービス アカウント権限を付与して GKE にデプロイします。

```sh
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$(gcloud projects describe $PROJECT_ID \
    --format="value(projectNumber)")-compute@developer.gserviceaccount.com \
    --role="roles/container.admin"
```

    この IAM ロールの詳細については、[container.admin](https://cloud.google.com/iam/docs/understanding-roles?hl=ja#kubernetes-engine-roles) ロールをご覧ください。

  - Cloud Build サービス アカウントに、Google Cloud Deploy オペレーションの呼び出しに必要な権限を付与します。

```sh
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$(gcloud projects describe $PROJECT_ID \
    --format="value(projectNumber)")@cloudbuild.gserviceaccount.com \
    --role="roles/iam.serviceAccountUser"
```

    Cloud Build は、Google Cloud Deploy を呼び出すときに、Compute Engine サービス アカウントを使用してリリースを作成します。そのため、この権限が必要になります。

    この IAM ロールの詳細については、[iam.serviceAccountUser](https://cloud.google.com/compute/docs/access/iam?hl=ja#the_serviceaccountuser_role) ロールをご覧ください。

これで、CI / CD パイプラインに必要な権限が付与されました。

### **GKE クラスタを作成する**
このセクションでは、ステージング環境と本番環境(どちらも GKE クラスタ)を作成します(minikube を使用するため、ここで開発クラスタの設定を行う必要はありません)。

1. ステージング環境と本番環境用の GKE クラスタを作成します。

```sh
gcloud container clusters create-auto staging \
    --region us-central1 \
    --project=$(gcloud config get-value project) \
    --async
```
```sh
gcloud container clusters create-auto prod \
    --region us-central1 \
    --project=$(gcloud config get-value project) \
    --async
```

  ステージング クラスタでは、コードの変更をテストします。ステージング環境のデプロイがアプリケーションに悪影響を及ぼさないことを確認したら、本番環境にデプロイします。

2. 次のコマンドを実行して、ステージング環境のクラスタと本番環境クラスタの両方の出力が STATUS: RUNNING であることを確認します。

```sh
gcloud container clusters list
```

3. ステージング環境のクラスタと本番環境クラスタの kubeconfig ファイルの認証情報を取得します。

    これらの認証情報を使用して GKE クラスタと情報を交換します。たとえば、アプリケーションが正しく実行されているかどうかを確認します。

```sh
gcloud container clusters get-credentials staging --region us-central1
```
```sh
gcloud container clusters get-credentials prod --region us-central1
```

ステージング環境と本番環境の GKE クラスタが作成されました。

### **IDE を開いてリポジトリのクローンを作成する**

リポジトリのクローンを作成して、開発環境でアプリケーションを表示するには、次の操作を行います。

1. [リポジトリのクローンを作成し、Cloud Shell で開きます](https://ssh.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https%3A%2F%2Fgithub.com%2Fgoogle%2Fgolden-path-for-app-delivery&cloudshell_git_branch=main&cloudshell_open_in_editor=README.md&hl=ja)。

2. [Confirm] をクリックします。

    Cloud Shell エディタが開き、サンプル リポジトリのクローンが作成されます。

    Cloud Shell エディタでアプリケーションのコードを表示できるようになりました。

3. このチュートリアルで使用するプロジェクトを指定します。

```sh
export PROJECT_ID={{project-id}}
gcloud config set project $PROJECT_ID
```

    ダイアログが表示された場合は、[承認] をクリックします。

これで、開発環境にアプリケーションのソースコードが作成されました。

このソース リポジトリには、CI / CD パイプラインに必要な Cloud Build ファイルと Google Cloud Deploy ファイルが含まれています。

### **ソースコード用のリポジトリとコンテナ用のリポジトリを作成する**

このセクションでは、ソースコード用のリポジトリを Cloud Source Repositories に設定し、CI / CD パイプラインによってビルドされたコンテナを格納する Artifact Registry にリポジトリを設定します。

1. Cloud Source Repositories で、ソースコードを格納するリポジトリを作成し、CI / CD プロセスにリンクします。

```sh
gcloud source repos create cicd-sample
```

2. Google Cloud Deploy 構成のターゲットが適切なプロジェクトであることを確認します。

```sh
sed -i s/project-id-placeholder/$(gcloud config get-value project)/g deploy/*
git config --global credential.https://source.developers.google.com.helper gcloud.sh
git remote add google https://source.developers.google.com/p/$(gcloud config get-value project)/r/cicd-sample
```

3. ソースコードをリポジトリに push します。

```sh
git push --all google
```

4. Artifact Registry にイメージ リポジトリを作成します。

```sh
gcloud artifacts repositories create cicd-sample-repo \
    --repository-format=Docker \
    --location us-central1
```

これで、Cloud Source Repositories にソースコードのリポジトリが作成され、Artifact Registry にアプリケーション コンテナのリポジトリが作成されました。Cloud Source Repositories リポジトリを使用すると、ソースコードのクローンを作成して CI / CD パイプラインに接続できます。

## **CI / CD パイプラインを構成する**

このセクションでは、アプリケーション オペレータとして機能し、CI/CD パイプラインを構成します。パイプラインは、CI の場合は Cloud Build、CD の場合は Google Cloud Deploy を使用します。パイプラインのステップは、Cloud Build トリガーで定義されます。

### 1. Cloud Build 用 Cloud Storage バケット作成

Cloud Build 用の Cloud Storage バケットを作成して `artifacts.json` ファイル(Skaffold によってビルドごとに生成されたアーティファクトを追跡)を保存します。

```sh
gsutil mb gs://$(gcloud config get-value project)-gceme-artifacts/
```

トレースを簡単に行えるため、各ビルドの `artifacts.json` ファイルを 1 か所に保存することをおすすめします。これにより、トラブルシューティングが容易になります。

### 2. `cloudbuild.yaml` ファイル確認

`cloudbuild.yaml` ファイルを確認します。これは Cloud Build トリガーを定義し、クローン作成したソース リポジトリですでに構成されています。

このファイルで、ソースコード リポジトリのメインブランチに対して新しい push が行われるたびに呼び出されるトリガーが定義されます。

  CI / CD パイプラインの手順は、このファイルで定義されています。

- Cloud Build は、Skaffold を使用してアプリケーション コンテナをビルドします。

- Cloud Build がビルドの `artifacts.json` ファイルを Cloud Storage バケットに配置します。

- Cloud Build がアプリケーション コンテナを Artifact Registry に配置します。

- Cloud Build がアプリケーション コンテナでテストを実行します。

- `gcloud beta deploy apply` コマンドは、次のファイルを Google Cloud Deploy サービスに登録します。

  - 配信パイプラインである `deploy/pipeline.yaml`
  - ターゲット ファイルである `deploy/staging.yaml` と `deploy/prod.yaml`

      ファイルが登録されると、Google Cloud Deploy はパイプラインとターゲットが存在しない場合は作成し、構成が変更された場合には再作成します。ターゲットは、ステージング環境と本番環境です。

- Google Cloud Deploy では、デリバリー パイプライン用の新しいリリースが作成されます。

  このリリースでは、CI プロセスでビルドとテストが行われたアプリケーション コンテナが参照されています。

-  Google Cloud Deploy により、リリースがステージング環境にデプロイされます。

デリバリー パイプラインとターゲットは Google Cloud Deploy によって管理され、ソースコードから切り離されます。この分離により、アプリケーションのソースコードが変更されたときに配信パイプラインとターゲット ファイルを更新する必要がなくなります。

### 3. Cloud Build トリガーを作成する

```sh
gcloud beta builds triggers create cloud-source-repositories \
    --name="cicd-sample-main" \
    --repo="cicd-sample" \
    --branch-pattern="main" \
    --build-config="cloudbuild.yaml"
```

このトリガーは、Cloud Build にソース リポジトリを監視するように指示し、`cloudbuild.yaml` ファイルを使用してリポジトリに対する変更に対応します。このトリガーは、メインブランチに新しい push があるたびに呼び出されます。

### 4. ビルドがないことを確認
[Cloud Build](https://console.cloud.google.com/cloud-build/dashboard?hl=ja&_ga=2.241990337.2113702237.1667787342-853816604.1666918848) に移動して、アプリケーションのビルドがないことを確認します。

これで CI パイプラインと CD パイプラインが設定され、リポジトリのメインブランチでトリガーが作成されました。

## **デベロッパー ワークスペース内でアプリケーションを変更する**

このセクションでは、アプリケーション デベロッパーとして作業します。

アプリケーションを開発する際には、開発ワークスペースとして Cloud Code を使用し、アプリケーションの変更と検証を繰り返します。

- アプリケーションに変更を加えます。
- 新しいコードをビルドしてテストします。
- アプリケーションを minikube クラスタにデプロイし、ユーザー向けの変更を検証します。
- メイン リポジトリに変更を送信します。

この変更がメイン リポジトリに commit されると、Cloud Build トリガーが CI / CD パイプラインを開始します。

### アプリケーションをビルドし、テストして、実行する

このセクションでは、アプリケーションをビルド、テスト、デプロイしてアクセスします。

前のセクションで使用したのと同じ Cloud Shell エディタ インスタンスを使用します。エディタを閉じた場合は、ブラウザで [ide.cloud.google.com](https://ide.cloud.google.com/?hl=ja) に移動して Cloud Shell エディタを開きます。

#### 1. ターミナルで、minikube を起動します。

```sh
minikube start
```

minikube で Cloud Shell にローカル Kubernetes クラスタを設定します。この設定が完了するまでに数分かかります。完了すると、minikube プロセスは Cloud Shell インスタンスのバックグラウンドで実行されます。

#### 2. Cloud Shell エディタの下部にあるペインで、**[Cloud Code]** を選択します。

#### 3. ターミナルとエディタの間に表示されるシンパネルで、**[Kubernetes で実行]** を選択します。

「`Use current context (minikube) to run the app?`」というプロンプトが表示されたら、[**はい**] をクリックします。

このコマンドにより、ソースコードがビルドされ、テストが実行されます。この処理には数分かかることがあります。テストには、単体テストと、デプロイ環境に設定されたルールを確認する事前構成された検証ステップが含まれます。これにより、開発環境で実行している場合でもデプロイの問題を警告できます。

[**出力**] タブには、Skaffold がアプリケーションをビルドしてデプロイする際の進行状況が表示されます。

このセクションを通してこのタブは開いたままにします。

ビルドとテストが終了すると、[**出力**] タブに `Update succeeded` と表示され、2 つの URL が表示されます。

アプリをビルドしてテストすると、Cloud Code の [**出力**] タブにログと URL がストリーミングされます。開発環境で変更を行い、テストを実行すると、開発環境のアプリのバージョンが表示され、正常に動作していることを確認できます。

出力には `Watching for changes...` も表示されます。これは、ウォッチモードが有効になっていることを意味します。Cloud Code がウォッチモードになっている間、このサービスはリポジトリに保存されている変更を検出し、最新の変更を使用してアプリを自動的に再ビルドして再デプロイします。

#### 4. Cloud Code ターミナルに出力された最初の URL(`http://localhost:8080`)にポインタを合わせます。

表示されたツールチップで、[**Open Web Preview**] を選択します。

Cloud Code は、バックグラウンドの minikube で実行されている `cicd-sample` サービスにトラフィックを自動的にポート転送をします。

ブラウザでページを更新します。

[**カウンタ**] の横にある数字が増加し、アプリが更新に応答していることを示します。

ローカル環境で変更を加える際にアプリケーションを表示できるように、ブラウザでこのページを開いたままにします。

これで、開発環境でのアプリケーションのビルドとテストが完了しました。アプリケーションを minikube で実行している開発クラスタにデプロイし、アプリケーションのユーザー向けの動作を確認しました。

### 変更する

このセクションでは、アプリケーションに変更を加えて、開発クラスタでのアプリの実行に合わせて変更を表示します。

#### 1. Cloud Shell エディタで、`index.html` ファイルを開きます。

#### 2. 文字列 `Sample App Info` を検索し、タイトルで小文字が使用されるように `sample app info` に変更します。

ファイルは自動的に保存され、アプリケーション コンテナの再構築がトリガーされます。

Cloud Code が変更を検出し、自動的に再デプロイします。[**出力**] タブに `Update initiated` が表示されます。この再デプロイが完了するまでに数分かかります。

この自動再デプロイ機能は、Kubernetes クラスタで実行されている任意のアプリケーションで使用できます。

ビルドが完了したら、アプリを開いているブラウザに移動して、ページを更新します。

更新すると、テキストで小文字が使用されるようになります。

この設定により、どのようなアーキテクチャのどのコンポーネントでも、自動的に再読み込みが行われます。Cloud Code と minikube を使用すると、Kubernetes で実行されているものすべてに、このホットコード リロード機能が備わっています。

Cloud Code では、Kubernetes クラスタにデプロイされたアプリケーションをデバッグできます。これらの手順はこのチュートリアルでは扱いませんが、[Kubernetes アプリケーションのデバッグ](https://cloud.google.com/code/docs/shell/debug?hl=ja)をご覧ください。

### コードを commit する

アプリケーションに変更を加えたので、コードを commit します。

#### 1. Git ID を構成します。

```sh
git config --global user.email "YOU@EXAMPLE.COM"
git config --global user.name "NAME"
```

次のように置き換えます。

- ***YOU@EXAMPLE.COM*** は、GitHub アカウントに接続されているメールアドレスに置き換えます。
- ***NAME*** は、GitHub アカウントに接続されている名前に置き換えます。

#### 2. ターミナルからコードを commit します。

```sh
git add .
git commit -m "use lowercase for: sample app info"
```

ここでは、`git push` コマンドを実行する必要はありません。これは後で行われます。

開発環境での作業で、アプリケーションを変更して変更をビルドしてテストし、これらの変更のユーザー向けの動作を確認しました。開発環境のテストにはガバナンス チェックが含まれています。これにより、本番環境での問題を引き起こす問題を解決できます。

このチュートリアルでは、コードをメイン リポジトリに commit しますが、コードのレビューは行いません。ただし、ソフトウェア開発では、コードレビューまたは変更承認が推奨されます。

変更承認のベスト プラクティスについて詳しくは、[変更承認の効率化](https://cloud.google.com/architecture/devops/devops-process-streamlining-change-approval?hl=ja)をご覧ください。

## **本番環境に変更をデプロイする**

このセクションでは、アプリケーション オペレーターとして、次の処理を行います。

- リリースをステージング環境にデプロイする CI / CD パイプラインをトリガーします。
- 本番環境へリリースを昇格させて承認します。

### CI / CD パイプラインを開始してステージング環境にデプロイする

このセクションでは、Cloud Build トリガーを呼び出して CI/CD パイプラインを開始します。このトリガーは、変更がメイン リポジトリに commit されるたびに呼び出されます。手動トリガーを使用して CI システムを開始することもできます。

#### 1. Cloud Shell エディタで、次のコマンドを実行してビルドをトリガーします。

```sh
git push google
```

このビルドには、`cicd-sample` に加えた変更が含まれています。

#### 2. [Cloud Build ダッシュボード](https://console.cloud.google.com/cloud-build/dashboard?hl=ja&_ga=2.237074371.2113702237.1667787342-853816604.1666918848)に戻り、ビルドが作成されたことを確認します。

#### 4. 右側のビルドログで [**Running: cicd-sample - cicd-sample-main**] をクリックし、各ステップの開始と終了を示す青いテキストを探します。

**ステップ 0** は、`cloudbuild.yaml` ファイルからの `skaffold build` 手順と `skaffold test` 手順の出力を示します。**ステップ 0**（パイプラインの CI 部分）のビルドタスクとテストタスクに合格したため、**ステップ 1**（パイプラインの CD 部分）のデプロイタスクが実行されるようになりました。

このステップは正常に完了し、次のメッセージが表示されます。

`Created Google Cloud Deploy rollout ROLLOUT_NAME in target staging`

#### 4. [Google Cloud Deploy 配信パイプライン ページ](https://console.cloud.google.com/deploy/delivery-pipelines?hl=ja&_ga=2.132930509.2113702237.1667787342-853816604.1666918848)を開き、`cicd-sample delivery` パイプラインをクリックします。

アプリケーションはステージング環境にデプロイされますが、本番環境にはデプロイされません。

#### 5. アプリケーションがステージング環境で正常に動作していることを確認します。

```sh
kubectl proxy --port 8001 --context gke_$(gcloud config get-value project)_us-central1_staging
```

このコマンドにより、アプリケーションにアクセスするための kubectl プロキシが設定されます。

#### 6. Cloud Shell からアプリケーションにアクセスします。

- Cloud Shell エディタで、新しいターミナルタブを開きます。

- リクエストを localhost に送信して、カウンタをインクリメントします。

```sh
curl -s http://localhost:8001/api/v1/namespaces/default/services/cicd-sample:8080/proxy/ | grep -A 1 Counter
```

    このコマンドは複数回実行でき、毎回カウンタ値が増分するのを確認できます。

    アプリを表示すると、変更したテキストがステージング環境にデプロイしたバージョンのアプリケーションに含まれていることがわかります。

- 2 つ目のタブを閉じます。

- 最初のタブで `Control+C` を押して、プロキシを停止します。

Cloud Build トリガーを呼び出して CI プロセスを開始しました。これには、アプリケーションのビルド、ステージング環境へのデプロイ、ステージング環境でのアプリケーションの動作検証のテスト実行が含まれます。

コードのビルドとテストがステージング環境で合格すると、CI プロセスは成功します。その後、CI プロセスが成功すると、Google Cloud Deploy で CD システムが開始されます。

### リリースを本番環境に昇格させる

このセクションでは、ステージング環境から本番環境にリリースをプロモートします。本番環境ターゲットは承認を必要とするように事前に構成されているため、手動で承認します。

独自の CI/CD パイプラインの場合、本番環境への完全なデプロイを行う前に、デプロイを段階的に実行するデプロイ戦略を使用することをおすすめします。デプロイを段階的に実行すると、問題を簡単に検出でき、必要に応じて以前のリリースを復元できます。

リリースを本番環境に昇格させる方法は次のとおりです。

#### 1. [Google Cloud Deploy デリバリー パイプラインの概要](https://console.cloud.google.com/deploy/delivery-pipelines?hl=ja&_ga=2.136944627.2113702237.1667787342-853816604.1666918848)を開き、cicd-sample パイプラインを選択します。

#### 2. ステージング環境から本番環境にデプロイを昇格します。手順は次のとおりです。

- ページの上部にあるパイプライン図で、ステージング ボックスの青色の [昇格] ボタンをクリックします。

- 表示されたウィンドウで、下部にある [昇格] ボタンをクリックします。

  デプロイはまだ本番環境では実行されていません。必要な手動承認を待機しています。

#### 3. デプロイを手動で承認します。

  - パイプラインの可視化で、ステージング ボックスと本番環境ボックスの間の [**レビュー**] ボタンをクリックします。

  - 表示されたウィンドウで [**レビュー**] ボタンをクリックします。

   - 次のウィンドウで [**承認**] をクリックします。

- [Google Cloud Deploy デリバリー パイプラインの概要](https://console.cloud.google.com/deploy/delivery-pipelines?hl=ja&_ga=2.182088292.2113702237.1667787342-853816604.1666918848)に戻り、cicd-sample パイプラインを選択します。

#### 4. パイプラインの可視化で prod ボックスが緑色で表示されたら（ロールアウトが成功したことを意味する）、アプリケーションへのアクセスに使用する kubectl プロキシを設定して、アプリケーションが本番環境で動作することを確認します。

```sh
kubectl proxy --port 8002 --context gke_$(gcloud config get-value project)_us-central1_prod
```

#### 5. Cloud Shell からアプリケーションにアクセスします。

  - Cloud Shell エディタで、新しいターミナルタブを開きます。

  - カウンタをインクリメントします。

```sh
curl -s http://localhost:8002/api/v1/namespaces/default/services/cicd-sample:8080/proxy/ | grep -A 1 Counter
```

    このコマンドは複数回実行でき、毎回カウンタ値が増分するのを確認できます。

  - この 2 つ目のターミナルタブを閉じます。

  - 最初のタブで `Control+C` を押して、プロキシを停止します。

これで昇格して、本番環境へのデプロイが承認されました。最近変更したアプリケーションは、本番環境で動作するようになりました。

## クリーンアップ

このチュートリアルで使用したリソースについて、Google Cloud アカウントに課金されないようにするには、リソースを含むプロジェクトを削除するか、プロジェクトを維持して個々のリソースを削除します。

### オプション 1: プロジェクトを削除する

⚠️⚠️⚠️ **注意**: プロジェクトを削除すると、次のような影響があります。

  - **プロジェクト内のすべてのものが削除されます**。このチュートリアルで既存のプロジェクトを使用した場合、それを削除すると、そのプロジェクトで行った他の作業もすべて削除されます。

  - **カスタム プロジェクト ID が失われます**。このプロジェクトを作成したときに、将来使用するカスタム プロジェクト ID を作成した可能性があります。そのプロジェクト ID を使用した URL（たとえば、appspot.com）を保持するには、プロジェクト全体ではなくプロジェクト内の選択したリソースだけを削除します。

複数のチュートリアルとクイックスタートを検討する予定がある場合は、プロジェクトを再利用すると、プロジェクトの割り当て制限を超えないようにできます。

#### 1. Google Cloud コンソールで、[リソースの管理] ページに移動します。
[[リソースの管理] に移動](https://console.cloud.google.com/iam-admin/projects?hl=ja&_ga=2.182220388.2113702237.1667787342-853816604.1666918848)

#### 2. プロジェクト リストで、削除するプロジェクトを選択し、[削除] をクリックします。

#### 3. ダイアログでプロジェクト ID を入力し、[シャットダウン] をクリックしてプロジェクトを削除します。

### オプション 2: 個々のリソースを削除する

#### 1. Google Cloud Deploy パイプラインを削除します。

```sh
gcloud deploy delivery-pipelines delete cicd-sample --region=us-central1 --force
```

#### 2. Cloud Build トリガーを削除します。

```sh
gcloud beta builds triggers delete cicd-sample-main
```

#### 3. ステージング クラスタと本番環境クラスタを削除します。

```sh
gcloud container clusters delete staging --region us-central1
```
```sh
gcloud container clusters delete prod --region us-central1
```

#### 4. Cloud Source Repositories でリポジトリを削除します。

```sh
gcloud source repos delete cicd-sample
```

#### 5. Cloud Storage バケットを削除します。

```sh
gsutil rm -r gs://$(gcloud config get-value project)-gceme-artifacts/
```
```sh
gsutil rm -r gs://$(gcloud config get-value project)_clouddeploy/
```

#### 6. Artifact Registry のリポジトリを削除します。

```sh
gcloud artifacts repositories delete cicd-sample-repo \
    --location us-central1
```

## 次のステップ

- プライベート GKE インスタンスにデプロイする方法については、[Virtual Private Cloud ネットワークの限定公開クラスタへのデプロイ](https://cloud.google.com/deploy/docs/execution-environment?hl=ja#deploying_to_a_private_cluster_on_a_network)をご覧ください。

- デプロイの自動化に関するベスト プラクティスについては、以下をご覧ください。
  - [DevOps 技術: デプロイ自動化](https://cloud.google.com/architecture/devops/devops-tech-deployment-automation?hl=ja)。デプロイの自動化を実装、改善、測定する方法。
  - Architecture Framework からの[デプロイの自動化](https://cloud.google.com/architecture/framework/operational-excellence/automate-your-deployments?hl=ja)。
- デプロイ戦略の詳細については、以下をご覧ください。
  - アーキテクチャ フレームワークから[デプロイを段階的に実行する](https://cloud.google.com/architecture/framework/operational-excellence/automate-your-deployments?hl=ja#launch_deployments_gradually)。
  - [アプリケーションのデプロイとテストの戦略](https://cloud.google.com/architecture/application-deployment-and-testing-strategies?hl=ja)
  - [GKE でのデプロイとテストの戦略の実装](https://cloud.google.com/architecture/implementing-deployment-and-testing-strategies-on-gke?hl=ja)のチュートリアル。
- [Cloud アーキテクチャ センター](https://cloud.google.com/architecture?hl=ja)で、その他のリファレンス アーキテクチャ、図、チュートリアル、ベスト プラクティスを確認する。
