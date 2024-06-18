<!-- TODO: Add the instruction for setting up the environmental variables? -->

# [GIG ハンズオン] **Cloud Code、Cloud Build、Cloud Deploy、Cloud Run を使用したアプリの開発と配信**

本ハンズオンは、以下のドキュメントを元に作成されています。
[オリジナル公式ドキュメント](https://cloud.google.com/deploy/docs/deploy-app-run?hl=ja)

## 目次 - Table of Contents
- [解説] ハンズオンの内容と目的
- アーキテクチャの概要
- 目標
- 費用
- Google Cloud プロジェクトの選択
- 環境を準備する
  - 権限を設定する
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
- [Option] カナリアデプロイ戦略、複数ターゲットへのデプロイ、デプロイ後の確認の設定
- クリーンアップ
  - オプション1: プロジェクトを削除する
  - オプション2: 個々のリソースを削除する
- 次のステップ

## [解説] **ハンズオンの内容と目的**

このチュートリアルでは、Google Cloud のツール群を使用して、開発、継続的インテグレーション(CI)、継続的デリバリー(CD)システムを設定し、使用する方法について説明します。このシステムを使用すると、アプリケーションを開発し、[Cloud Run](https://cloud.google.com/run?hl=ja)にデプロイできます。

このチュートリアルは、ソフトウェア デベロッパーとオペレーターの両方を対象としており、完了すると両方の役割を果たせます。まず、CI / CD パイプラインを設定するオペレーターの役割を果たします。このパイプラインの主要なコンポーネントは、[Cloud Build](https://cloud.google.com/build?hl=ja)、[Artifact Registry](https://cloud.google.com/artifact-registry?hl=ja)、[Cloud Deploy](https://cloud.google.com/deploy?hl=ja) です。

次にデベロッパーとして、[Cloud Code](https://cloud.google.com/code?hl=ja) を使用してアプリケーションを変更します。すると、このパイプラインが提供する統合された環境を確認することができます。

最後に、オペレータとして、アプリケーションを本番環境にデプロイする手順を実行します。

このチュートリアルは、Google Cloud での gcloud コマンドの実行と Cloud Run へのアプリケーション コンテナのデプロイに精通していることを前提としています。

この統合システムの主な特徴は次のとおりです。

- **開発とデプロイが迅速になります。**

  デベロッパー ワークスペースで変更を検証できるため、開発ループが効率的です。自動化された CI/CD システムと環境間の差分を小さくすることにより、本番環境に変更を展開する際に、より多くの問題を検出できるため、デプロイが高速化されます。

- **開発、ステージング、本番環境全体で一貫性の向上によるメリットを享受できます。**

  このシステムのコンポーネントは、共通の Google Cloud ツールセットを使用します。

- **さまざまな環境で構成を再利用します。**

  Skaffold は異なる環境に対して共通の構成形式を利用可能であるため、再利用することができます。また、デベロッパーとオペレーターが同じ構成を更新して使用することもできます。

- **ワークフローの早い段階でガバナンスを適用します。**

  このシステムでは、本番環境、CI システム、開発環境でガバナンスに関する検証テストが適用されます。開発環境では、ガバナンスを適用することで、問題を検出し、早期に修正できます。

- **独自のツールでソフトウェア配信を管理します。**

  継続的デリバリーはフルマネージドであり、CD パイプラインのステージをレンダリングやデプロイの細部から分離します。

## **アーキテクチャの概要**

次の図に、このチュートリアルで使用するリソースを示します。

![](https://raw.githubusercontent.com/google-cloud-japan/gig-training-materials/main/gig08-03/img/diagram.png)

このパイプラインを構成する 3 つの主要コンポーネントは次のとおりです。

1. 開発ワークスペースとしての **Cloud Code**

    [Cloud Shell](https://cloud.google.com/shell?hl=ja) で Cloud Code を活用し、変更を確認できます。Cloud Shell は、ブラウザからアクセスできるオンライン開発環境です。コンピューティング リソース、メモリ、統合開発環境(IDE)を備え、Cloud Code もインストールされます。

2. アプリケーションのビルドとテストを行うための **Cloud Build**(パイプラインの「CI」部分)

    パイプラインのこの部分には、次のアクションが含まれます。

    - Cloud Build は、Cloud Build トリガーを使用して、ソース リポジトリに対する変更をモニタリングします。
    - メインブランチに変更が commit されると、Cloud Build トリガーは次の処理を行います。
      - アプリケーション コンテナを再構築します。
      - アプリケーション コンテナを Artifact Registry に配置します。
      - コンテナでテストを実行します。
      - Cloud Deploy を呼び出して、コンテナをステージング環境にデプロイします。このチュートリアルでは、ステージング環境は Cloud Run サービスです。
    - ビルドとテストが成功したら、Cloud Deploy を使用して、ステージング環境から本番環境にコンテナを昇格できます。

3. デプロイを管理するための **Cloud Deploy**(パイプラインの「CD」部分)

    パイプラインのこの部分では、Cloud Deploy が次の処理を行います。

    - [配信パイプライン](https://cloud.google.com/deploy/docs/terminology?hl=ja#delivery_pipeline)と[ターゲット](https://cloud.google.com/deploy/docs/terminology?hl=ja#target)を登録します。ターゲットはステージング環境と本番環境を表します。
    - Cloud Storage バケットを作成します。このバケットには Skaffold レンダリング ソースとレンダリングされたマニフェストが保存されます。
    - ソースコードを変更するたびに[新しいリリース](https://cloud.google.com/deploy/docs/terminology?hl=ja#release)を行います。このチュートリアルでは 1 つの変更を行い、新しいリリースを 1 つ行います。
    - アプリケーションを本番環境にデプロイします。この本番環境へのデプロイでは、運用担当者(または指定した人)がデプロイを手動で承認します。このチュートリアルでは、本番環境は Cloud Run サービスです。

Container & Kubernetes ネイティブ アプリケーションの継続的な開発を容易にするコマンドライン ツールである [Skaffold](https://skaffold.dev/) は、これらのコンポーネントの基盤であり、開発、ステージング、本番環境の間で構成を共有できます。

元となるアプリケーションのソースコードは GitHub にあります。このチュートリアルでは、このリポジトリのクローンを Cloud Source Repositories に作成して、CI / CD パイプラインに接続します。

このチュートリアルでは、システムのほとんどのコンポーネントで Google Cloud プロダクトを使用し、Skaffold でシステムを統合しています。Skaffold はオープンソースであるため、同様に Google Cloud、社内コンポーネント、サードパーティ コンポーネントを組み合わせることができます。このソリューションはモジュール方式を採用しているため、開発パイプラインとデプロイ パイプラインの一部として段階的に導入できます。

**注意**: このチュートリアルでは Git リポジトリとして Cloud Source Repositories を利用していますが、実際の開発では [GitHub, GitLab, Bitbucket を Cloud Build に接続](https://cloud.google.com/build/docs/repositories)して利用してください。

## **目標**

オペレーターとして、次の操作を行います。

- CI パイプラインと CD パイプラインを設定します。この設定には次のものが含まれます。
  - 必要な権限を設定する。
  - ソースコード用のリポジトリを Cloud Source Repositories に作成する。
  - アプリケーション コンテナ用のリポジトリを Artifact Registry に作成する。
  - メインの GitHub リポジトリに Cloud Build トリガーを作成する。
  - Cloud Deploy デリバリー パイプラインとターゲットを作成する。ターゲットはステージング環境と本番環境です。
- ステージング環境にデプロイする CI / CD プロセスを開始し、本番環境に昇格させる。

開発者として、アプリケーションに変更を加えます。手順は次のとおりです。

- 事前構成された開発環境と連携するように、リポジトリのクローンを作成する。
- デベロッパー ワークスペース内でアプリケーションを変更する
- 変更をビルドおよびテストします。テストには、ガバナンスの検証テストが含まれます。
- 開発環境 (Cloud Shell のローカル環境) で変更を表示して検証します。
- メイン リポジトリに変更を commit する。

## **費用**

このチュートリアルでは、課金対象である次の Google Cloud コンポーネントを使用します。

[Cloud Build](https://cloud.google.com/build/pricing?hl=ja)
[Cloud Deploy](https://cloud.google.com/deploy/pricing?hl=ja)
[Artifact Registry](https://cloud.google.com/artifact-registry/pricing?hl=ja)
[Cloud Run](https://cloud.google.com/run/pricing?hl=ja)
[Cloud Source Repositories](https://cloud.google.com/source-repositories/pricing?hl=ja)
[Cloud Storage](https://cloud.google.com/storage/pricing?hl=ja)
[料金計算ツール](https://cloud.google.com/products/calculator?hl=ja)を使うと、予想使用量に基づいて費用の見積もりを生成できます。

このチュートリアルを終了した後、作成したリソースを削除すると、それ以上の請求は発生しません。詳細については、[クリーンアップ](https://cloud.google.com/deploy/docs/deploy-app-run?hl=ja#clean-up)をご覧ください。

## Google Cloud プロジェクトの選択

1. ハンズオンを行う Google Cloud プロジェクトを作成し、 Google Cloud プロジェクトを選択して **Start/開始** をクリックしてください。

**なるべく新しいプロジェクトを作成してください。**

<walkthrough-project-setup>
</walkthrough-project-setup>

2. Cloud プロジェクトに対して課金が有効になっていることを確認します。詳しくは、[プロジェクトで課金が有効になっているかどうかを確認する方法](https://cloud.google.com/billing/docs/how-to/verify-billing-enabled?hl=ja)をご覧ください。

3. Artifact Registry, Cloud Build, Cloud Deploy, Cloud Source Repositories, Cloud Run, Resource Manager, Service Networking API を有効にします。

    [API を有効にするリンク](https://console.cloud.google.com/flows/enableapi?apiid=artifactregistry.googleapis.com%2Ccloudbuild.googleapis.com%2Cclouddeploy.googleapis.com%2Csourcerepo.googleapis.com%2Crun.googleapis.com%2C+cloudresourcemanager.googleapis.com%2Cservicenetworking.googleapis.com&%3Bredirect=https%3A%2F%2Fconsole.cloud.google.com&hl=ja)

4. Google Cloud コンソールで、「Cloud Shell をアクティブにする」をクリックします。

    [Cloud Shell をアクティブにする](https://console.cloud.google.com/?cloudshell=true&hl=ja)

    Google Cloud コンソールの下部で [Cloud Shell](https://cloud.google.com/shell/docs/how-cloud-shell-works?hl=ja) セッションが開始し、コマンドライン プロンプトが表示されます。Cloud Shell はシェル環境です。Google Cloud CLI がすでにインストールされており、現在のプロジェクトの値もすでに設定されています。セッションが初期化されるまで数秒かかることがあります。

## **環境を準備する**

このセクションでは、アプリケーション オペレーターとして、次の処理を行います。

- 必要な権限を設定する。
- ソース リポジトリのクローンを作成する
- ソースコード用のリポジトリを Cloud Source Repositories に作成する。
- コンテナ アプリケーション用のリポジトリを Artifact Registry に作成します。

### **権限を設定する**

このセクションでは、CI / CD パイプラインの設定に必要なサービスアカウントを作成し、権限を付与します。

1. Cloud Shell エディタの新しいインスタンスで作業している場合は、このチュートリアルで使用するプロジェクトを指定します。

```sh
export PROJECT_ID=<walkthrough-project-id/>
gcloud config set project $PROJECT_ID
```

*PROJECT_ID* は、このチュートリアルで選択または作成したプロジェクトの ID ( <walkthrough-project-id/> )を設定します。

ダイアログが表示された場合は、[承認] をクリックします。

2. 必要なサービス アカウントを作成し、必要な権限を付与します。

  - まず必要なサービスアカウントを作成します。このハンズオンでは、ビルドを実行、ビルドしたコンテナイメージをデプロイ、コンテナイメージを実行の3つの要素があるので、それぞれにサービスアカウントを作成します。

  ```sh
  gcloud iam service-accounts create builder
  gcloud iam service-accounts create deployer
  gcloud iam service-accounts create runner
  ```

  参考: [サービスアカウントを指定しない場合に Cloud Build がデフォルトで利用するサービスアカウントは条件により異なります](https://cloud.google.com/build/docs/cloud-build-service-account-updates?hl=ja#org-policy)。このハンズオンでは条件によらず一定の動作をさせるため、また最小権限の原則に則って専用のサービスアカウントを利用しています。

  - Cloud Build がビルドを実行し、ログやビルドしたイメージを書き込むために必要なロールを付与します。

    ```sh
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member=serviceAccount:builder@$PROJECT_ID.iam.gserviceaccount.com \
        --role="roles/cloudbuild.builds.builder"
    ```

  - Cloud Build から Cloud Deploy の配信パイプラインとターゲットの定義を更新するためのロールを付与します。

    ```sh
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member=serviceAccount:builder@$PROJECT_ID.iam.gserviceaccount.com \
        --role="roles/clouddeploy.operator"
    ```

    この IAM ロールの詳細については、[clouddeploy.operator](https://cloud.google.com/deploy/docs/iam-roles-permissions?hl=ja#predefined_roles) ロールをご覧ください。

  - Cloud Build サービス アカウントに、Cloud Deploy オペレーションの呼び出しに必要な権限を付与します。

    ```sh
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member=serviceAccount:builder@$PROJECT_ID.iam.gserviceaccount.com \
        --role="roles/iam.serviceAccountUser"
    ```

    Cloud Build は、Cloud Deploy を呼び出すときに、別のサービス アカウント（`deployer@`）を使用してリリースを作成します。そのため、この権限が必要になります。

    この IAM ロールの詳細については、[iam.serviceAccountUser](https://cloud.google.com/compute/docs/access/iam?hl=ja#the_serviceaccountuser_role) ロールをご覧ください。

  - Cloud Deploy がジョブを実行するために必要なロールを付与します。

    ```sh
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member=serviceAccount:deployer@$PROJECT_ID.iam.gserviceaccount.com \
        --role="roles/clouddeploy.jobRunner"
    ```

  - Cloud Deploy が Cloud Run にデプロイするための管理ロールを付与します。

    ```sh
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member=serviceAccount:deployer@$PROJECT_ID.iam.gserviceaccount.com \
        --role="roles/run.admin"
    ```

    この IAM ロールの詳細については、[run.admin](https://cloud.google.com/iam/docs/understanding-roles?hl=ja#cloud-run-roles) ロールをご覧ください。

  - Cloud Deploy が Cloud Run へワークロードをデプロイするとき、間接的に `runner@` のサービスアカウントを利用するための権限を付与します。

    ```sh
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member=serviceAccount:deployer@$PROJECT_ID.iam.gserviceaccount.com \
        --role="roles/iam.serviceAccountUser"
    ```

これで、CI / CD パイプラインに必要な権限が付与されました。

### **IDE を開いてソースコードのあるディレクトリに移動する**

今回は既にクローンしたリポジトリにソースコードが含まれていますので、該当のディレクトリに移動します。

1. ディレクトリを移動する

    ```sh
    cd ~/cloudshell_open/gig-training-materials/gig08-03
    ```

このソース リポジトリには、CI / CD パイプラインに必要な Cloud Build ファイルと Cloud Deploy ファイルが含まれています。

### **ソースコード用のリポジトリとコンテナ用のリポジトリを作成する**

このセクションでは、ソースコード用のリポジトリを Cloud Source Repositories に設定し、CI / CD パイプラインによってビルドされたコンテナを格納する Artifact Registry にリポジトリを設定します。

1. Cloud Source Repositories で、ソースコードを格納するリポジトリを作成し、CI / CD プロセスにリンクします。

    ```sh
    gcloud source repos create cicd-sample
    ```

2. Git のリモートリポジトリを Cloud Source Repositories に設定します。

    ```sh
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

このセクションでは、アプリケーション オペレータとして機能し、CI/CD パイプラインを構成します。パイプラインでは、CI に Cloud Build、CD に Cloud Deploy を使用します。パイプラインのステップは、Cloud Build ビルド構成ファイルに記載します。

### 1. `cloudbuild.yaml` ファイル確認

Cloud Build ビルド構成ファイルの `cloudbuild.yaml` を確認します。クローンしたソース リポジトリにすでに含まれています。

このファイルで、ソースコード リポジトリのメインブランチに対して新しい push が行われるたびに呼び出されるビルド ステップが定義されます。

CI / CD パイプラインの手順は、このファイルで定義されています。

- Cloud Build は、コンテナイメージをビルドし、Artifact Registry に配置します。

- `gcloud deploy apply` コマンドは、 clouddeploy.yaml に定義された配信パイプライン、ターゲットを Cloud Deploy サービスに登録します。

ファイルが登録されると、Cloud Deploy はパイプラインとターゲットが存在しない場合は作成し、構成が変更された場合には再作成します。ターゲットは、ステージング環境と本番環境です。

- Cloud Deploy では、デリバリー パイプライン用の新しいリリースが作成されます。

  このリリースでは、CI プロセスでビルドとテストが行われたアプリケーション コンテナが参照されています。

-  Cloud Deploy により、リリースがステージング環境にデプロイされます。

デリバリー パイプラインとターゲットは Cloud Deploy によって管理され、ソースコードから切り離されます。この分離により、アプリケーションのソースコードが変更されたときに配信パイプラインとターゲット ファイルを更新する必要がなくなります。

### 2. Cloud Build トリガーを作成する

```sh
gcloud builds triggers create cloud-source-repositories \
    --name="cicd-sample-main" \
    --repo="cicd-sample" \
    --branch-pattern="main" \
    --build-config="gig08-03/cloudbuild.yaml" \
    --service-account="projects/$PROJECT_ID/serviceAccounts/builder@$PROJECT_ID.iam.gserviceaccount.com"
```

このトリガーは、Cloud Build にソース リポジトリを監視するように指示し、`cloudbuild.yaml` ファイルを使用してリポジトリに対する変更に対応します。このトリガーは、メインブランチに新しい push があるたびに呼び出されます。

### 3. ビルドがないことを確認
[Cloud Build](https://console.cloud.google.com/cloud-build/dashboard?hl=ja) に移動して、アプリケーションのビルドがないことを確認します。

これで CI パイプラインと CD パイプラインが設定され、リポジトリのメインブランチでトリガーが作成されました。

## **デベロッパー ワークスペース内でアプリケーションを変更する**

このセクションでは、アプリケーション デベロッパーとして作業します。

アプリケーションを開発する際には、開発ワークスペースとして Cloud Code を使用し、アプリケーションの変更と検証を繰り返します。

- アプリケーションに変更を加えます。
- 新しいコードをビルドします。
- アプリケーションを Cloud Run Emulator にデプロイし、ユーザー向けの変更を検証します。
- メイン リポジトリに変更を送信します。

この変更がメイン リポジトリに commit されると、Cloud Build トリガーが CI / CD パイプラインを開始します。

### アプリケーションをビルドし、実行する

このセクションでは、アプリケーションをビルド、テスト、デプロイしてアクセスします。

前のセクションで使用したのと同じ Cloud Shell エディタ インスタンスを使用します。エディタを閉じた場合は、ブラウザで [ide.cloud.google.com](https://ide.cloud.google.com/?hl=ja) に移動して Cloud Shell エディタを開きます。

#### 1. Cloud Shell エディタの下部にあるステータスバーで、**[Cloud Code]** を選択します。

#### 2. エディタの上部に表示される選択肢から、**[Run on Cloud Run Emulator]** を選択します。

ビルド設定のページが開くので、デフォルトのまま **[Run]** ボタンを選択すると、ソースコードがビルドされエミュレータ上で起動します。

[**出力**] タブには、Skaffold がアプリケーションをビルドしてデプロイする際の進行状況が表示されます。この処理には数分かかります。

進んでいないように見える場合、 `Cloud Run: Run/Debug Locally - Detailed` を選択すると、詳細なログを表示することができます。`Listening on port 8080` と表示されたら完了なので、元の `Cloud Run: Run/Debug Locally` に戻してください。

このセクションを通してこのタブは開いたままにします。

ビルドとテストが終了すると、[**出力**] タブに `Deploy completed` と表示され、URL が表示されます。

アプリをビルドすると、Cloud Code の [**出力**] タブにログと URL がストリーミングされます。開発環境で変更を行い、テストを実行すると、開発環境のアプリのバージョンが表示され、正常に動作していることを確認できます。

出力には `Watching for changes...` も表示されます。これは、ウォッチモードが有効になっていることを意味します。Cloud Code がウォッチモードになっている間、このサービスはリポジトリ内のファイルの変更を検出し、自動的にアプリの再ビルドと再デプロイを行います。

#### 3. Cloud Code ターミナルに出力された最初の URL(`http://localhost:8080`)にポインタを合わせます。

表示された [Follow link] のツールチップに沿って URL を開きます。

Cloud Code は、バックグラウンドの Cloud Run Emulator で実行されている `cicd-sample` サービスにトラフィックを自動的にポート転送します。

ローカル環境で変更を加える際にアプリケーションを表示できるように、ブラウザでこのページを開いたままにします。

これで、開発環境でのアプリケーションのビルドとテストが完了しました。アプリケーションを minikube 上の Cloud Run Emulator で実行している開発クラスタにデプロイし、アプリケーションのユーザー向けの動作を確認しました。

### 変更する

このセクションでは、アプリケーションに変更を加えて、開発クラスタでのアプリの実行に合わせて変更を表示します。

#### 1. Cloud Shell エディタで、`index.html` ファイルを開きます。

#### 2. 文字列 `Hello World!` を検索し、タイトルを `Hello GIG!` に変更します。

ファイルは自動的に保存され、アプリケーション コンテナの再構築がトリガーされます。

Cloud Code が変更を検出し、自動的に再デプロイします。[**出力**] タブに `Update initiated` が表示されます。この再デプロイが完了するまでに数分かかります。

ビルドが完了したら、アプリを開いているブラウザに移動して、ページを更新します。

更新すると、テキストが変更されています。

デフォルトの設定により、どのようなアーキテクチャのどのコンポーネントでも、自動的に再ビルドが行われます。

Cloud Code では、Cloud Run Emulator にデプロイされたアプリケーションをデバッグできます。これらの手順はこのチュートリアルでは扱いませんが、[Cloud Code for Cloud Shell で Cloud Run サービスをデバッグする](https://cloud.google.com/code/docs/shell/debug-service?hl=ja)をご覧ください。

### コードを commit する

アプリケーションに変更を加えたので、コードを commit します。
[Terminal] にタブを切り替えてください。

#### 1. Git ID を構成します。

```sh
git config --global user.email "YOU@EXAMPLE.COM"
git config --global user.name "NAME"
```

次のように置き換えます。

- ***YOU@EXAMPLE.COM*** は、メールアドレスに置き換えます。
- ***NAME*** は、名前に置き換えます。

#### 2. ターミナルからコードを commit します。

```sh
git add .
git commit -m "Hello GIG"
```

ここでは、`git push` コマンドを実行する必要はありません。これは後で行います。

開発環境での作業で、アプリケーションを変更して変更をビルドしてテストし、これらの変更のユーザー向けの動作を確認しました。開発環境のテストにはガバナンス チェックが含まれています。これにより、本番環境での問題を引き起こす問題を解決できます。

このチュートリアルでは、コードをメイン リポジトリに commit しますが、コードのレビューは行いません。ただし、ソフトウェア開発では、コードレビューまたは変更承認が推奨されます。

変更承認のベスト プラクティスについて詳しくは、[変更承認の効率化](https://cloud.google.com/architecture/devops/devops-process-streamlining-change-approval?hl=ja)をご覧ください。

## **本番環境に変更をデプロイする**

このセクションでは、アプリケーション オペレーターとして、次の処理を行います。

- リリースをステージング環境にデプロイする CI / CD パイプラインをトリガーします。
- 本番環境へリリースを昇格させて承認します。

### CI / CD パイプラインを開始してステージング環境にデプロイする

このセクションでは、Cloud Build トリガーを呼び出して CI/CD パイプラインを開始します。このトリガーは、変更がメイン リポジトリに commit されるたびに呼び出されます。手動トリガーを使用して CI システムを起動することもできます。

#### 1. Cloud Shell エディタで、次のコマンドを実行してビルドをトリガーします。

```sh
git push google
```

このビルドには、`cicd-sample` に加えた変更が含まれています。

#### 2. [Cloud Build ダッシュボード](https://console.cloud.google.com/cloud-build/dashboard?hl=ja)に戻り、ビルドが作成されたことを確認します。

#### 3. 右側のビルドログで [**Running: cicd-sample - cicd-sample-main**] をクリックし、各ステップの開始と終了を示す青いテキストを探します。

**ステップ 0** は、`cloudbuild.yaml` ファイルからの コンテナをビルドする手順の出力を示します。**ステップ 0**（パイプラインの CI 部分）のビルドタスクに合格したため、**ステップ 2, 3**（パイプラインの CD 部分）のデプロイタスクが実行されるようになりました。

このステップは正常に完了し、次のメッセージが表示されます。

`Created Cloud Deploy release rel-ab12345.`

#### 4. [Cloud Deploy デリバリー パイプライン ページ](https://console.cloud.google.com/deploy/delivery-pipelines?hl=ja)を開き、`cicd-sample` パイプラインをクリックします。

アプリケーションはステージング環境にデプロイされますが、本番環境にはデプロイされません。

#### 5. Cloud Run のステージング環境をアクセス可能にします。

Cloud Run サービスを公開するため、以下のコマンドを実行します。

```sh
gcloud run services set-iam-policy deploy-qs-dev policy.yaml --region=us-central1
```

`Replace existing policy (Y/n)?` と確認された場合は `y` で進めてください。

#### 6. アプリケーションがステージング環境で正常に動作していることを確認します。

[Cloud Run のコンソール画面](https://console.cloud.google.com/run?hl=ja) を開き、ステージング環境のアプリが正しく動作していることを確認します。

アプリを表示すると、変更したテキストがステージング環境にデプロイしたバージョンのアプリケーションに含まれていることがわかります。

- 2 つ目のタブを閉じます。

Cloud Build トリガーを呼び出して CI プロセスを開始しました。これには、アプリケーションのビルド、ステージング環境へのデプロイが含まれます。

コードのビルドがステージング環境で合格すると、CI プロセスは成功します。その後、CI プロセスが成功すると、Cloud Deploy で CD システムが開始されます。

### リリースを本番環境に昇格させる

このセクションでは、ステージング環境から本番環境にリリースをプロモートします。本番環境ターゲットは承認を必要とするように構成されているため、手動で承認します。

実際の CI/CD パイプラインの場合、本番環境への完全なデプロイを行う前に、デプロイを段階的に実行するデプロイ戦略を使用することをおすすめします。デプロイを段階的に実行すると、問題を簡単に検出でき、必要に応じて以前のリリースを復元できます。

リリースを本番環境に昇格させる方法は次のとおりです。

#### 1. [Cloud Deploy デリバリー パイプラインの概要](https://console.cloud.google.com/deploy/delivery-pipelines?hl=ja)を開き、cicd-sample パイプラインを選択します。

#### 2. ステージング環境から本番環境にデプロイを昇格します。手順は次のとおりです。

- ページの上部にあるパイプライン図で、ステージング ボックスの青色の [プロモート] ボタンをクリックします。

- 表示されたウィンドウで、下部にある [プロモート] ボタンをクリックします。

  デプロイはまだ本番環境では実行されていません。必要な手動承認を待機しています。

#### 3. デプロイを手動で承認します。

  - パイプラインの可視化で、ステージング ボックスと本番環境ボックスの間の [**確認**] ボタンをクリックします。

  - 表示されたウィンドウで [**レビュー**] ボタンをクリックします。

   - 次のウィンドウで [**承認**] をクリックします。

- [Cloud Deploy デリバリー パイプラインの概要](https://console.cloud.google.com/deploy/delivery-pipelines?hl=ja)に戻り、cicd-sample パイプラインを選択します。

#### 4. パイプラインの可視化で prod ボックスが緑色で表示されたら（ロールアウトが成功したことを意味する）、アプリケーションが本番環境で動作することを確認します。

Cloud Run サービスを公開するため、以下のコマンドを実行します。

```sh
gcloud run services set-iam-policy deploy-qs-prod policy.yaml --region=us-central1
```

[Cloud Run のコンソール画面](https://console.cloud.google.com/run?hl=ja) を開き、本番環境のアプリが正しく動作していることを確認します。

これで昇格して、本番環境へのデプロイが承認されました。最近変更したアプリケーションは、本番環境で動作するようになりました。

## [Option] 複数ターゲットへのデプロイ、カナリアデプロイ戦略、デプロイ後の確認の設定

Cloud Deploy では、複数のターゲットへのデプロイやカナリアデプロイ戦略がサポートされています（2023-06-26 GA）。またデプロイ結果を確認することもできます（2023-02-27 GA）。

なお、ここでは `clouddeploy.yaml` の設定内容のみを記載しています。詳細な手順は、以下を参照してください。

- [アプリを複数のターゲットに同時にデプロイする](https://cloud.google.com/deploy/docs/deploy-app-parallel?hl=ja)
- [アプリケーションをターゲットにカナリア デプロイする](https://cloud.google.com/deploy/docs/deploy-app-canary?hl=ja)
- [デプロイを確認する](https://cloud.google.com/deploy/docs/verify-deployment?hl=ja)

### 複数ターゲットへのデプロイ

`clouddeploy.yaml` を修正し、複数のターゲットへのデプロイを設定します（ `clouddeploy.yaml` にはコメントアウトされた状態で設定が含められています。修正する際はインデントに注意してください）。

ここでは、 asia-northeast1 にデプロイするターゲットと、マルチターゲットを追加して、マルチターゲットを指定するように `targetId` を修正しています。

ロールアウトの承認は子ターゲットではなくマルチターゲットに対して設定する必要があるので、`requireApproval: true` を `run-qsprod` から削除し、`run-qsprod-multi`へ移動します。

```yaml
# ...(略)...
serialPipeline:
  stages:
  - targetId: run-qsdev
    profiles: [dev]
  # - targetId: run-qsprod
  #   profiles: [prod]
  # multi target
  - targetId: run-qsprod-multi
    profiles: [prod]
    # ...(略)...
---
# ...(略)...
---
apiVersion: deploy.cloud.google.com/v1
kind: Target
metadata:
  name: run-qsprod-multi
description: production
requireApproval: true
multiTarget:
  targetIds: [run-qsprod, run-qsprod-tok]

---
apiVersion: deploy.cloud.google.com/v1
kind: Target
metadata:
  name: run-qsprod-tok
description: Cloud Run production service
run:
  location: projects/PROJECT_ID/locations/asia-northeast1
```

### カナリアデプロイ戦略

`clouddeploy.yaml` を修正し、カナリアデプロイ戦略を設定します（ `clouddeploy.yaml` にはコメントアウトされた状態で設定が含められています。修正する際はインデントに注意してください）。

なお、ここでは 10%, 25%, 50%, 100% (stable) の 4 つのフェーズが定義されています（100% は記述する必要はありません）。

```yaml
# ...(略)...
serialPipeline:
  stages:
  - targetId: run-qsdev
    profiles: [dev]
  - targetId: run-qsprod
    profiles: [prod]
    ## canary deployment
    strategy:
      canary:
        runtimeConfig:
          cloudRun:
            automaticTrafficControl: true
        canaryDeployment:
          percentages: [10, 25, 50]
          verify: false
```

### デプロイ後の確認

デプロイ後の確認を有効にするには、 `clouddeploy.yaml` のターゲットの設定に `strategy` 以下の3行を追記します。

```yaml
serialPipeline:
  stages:
    - targetId: run-qsdev
      profiles: [dev]
      strategy:
        standard:
          verify: true
```

実際の確認内容は `skaffold.yaml` に記載します。
今回のチュートリアルでは、HTML ファイルの見出しを `GIG` に書き換えたので、正しく書き換えられているかどうかを検証しています。`$CLOUD_RUN_SERVICE_URLS` には Cloud Run の URL が自動的に設定されます。他に利用可能な環境変数はドキュメントを参照してください。

```yaml
verify:
- name: verify-content-test
  container:
    name: curl
    image: curlimages/curl
    command: ["sh"]
    args: ["-c", "curl --silent $CLOUD_RUN_SERVICE_URLS | grep GIG"]
```

## クリーンアップ

このチュートリアルで使用したリソースについて、Google Cloud アカウントに課金されないようにするには、リソースを含むプロジェクトを削除するか、プロジェクトを維持して個々のリソースを削除します。

### オプション 1: プロジェクトを削除する

⚠️⚠️⚠️ **注意**: プロジェクトを削除すると、次のような影響があります。

  - **プロジェクト内のすべてのものが削除されます**。このチュートリアルで既存のプロジェクトを使用した場合、それを削除すると、そのプロジェクトで行った他の作業もすべて削除されます。

  - **カスタム プロジェクト ID が失われます**。このプロジェクトを作成したときに、将来使用するカスタム プロジェクト ID を作成した可能性があります。そのプロジェクト ID を使用した URL（たとえば、appspot.com）を保持するには、プロジェクト全体ではなくプロジェクト内の選択したリソースだけを削除します。

複数のチュートリアルとクイックスタートを検討する予定がある場合は、プロジェクトを再利用すると、プロジェクトの割り当て制限を超えないようにできます。

#### 1. Google Cloud コンソールで、[リソースの管理] ページに移動します。
[[リソースの管理] に移動](https://console.cloud.google.com/iam-admin/projects?hl=ja)

#### 2. プロジェクト リストで、削除するプロジェクトを選択し、[削除] をクリックします。

#### 3. ダイアログでプロジェクト ID を入力し、[シャットダウン] をクリックしてプロジェクトを削除します。

### オプション 2: 個々のリソースを削除する

#### 1. Cloud Deploy パイプラインを削除します。

```sh
gcloud deploy delivery-pipelines delete cicd-sample --region=us-central1 --force
```

#### 2. Cloud Build トリガーを削除します。

```sh
gcloud builds triggers delete cicd-sample-main
```

#### 3. ステージング 環境と本番環境を削除します。

```sh
gcloud run services delete deploy-qs-dev --region=us-central1
```
```sh
gcloud run services delete deploy-qs-prod --region=us-central1
```

#### 4. Cloud Source Repositories でリポジトリを削除します。

```sh
gcloud source repos delete cicd-sample
```

#### 5. Artifact Registry のリポジトリを削除します。

```sh
gcloud artifacts repositories delete cicd-sample-repo \
    --location us-central1
```

## Congratulations

<walkthrough-conclusion-trophy></walkthrough-conclusion-trophy>

お疲れ様でした！
