from time import time
from flask import Flask, render_template, request, Markup
from google.cloud import vision, storage
app = Flask(__name__)

# 今回使用するバケット名を設定
bucket_name = '<プロジェクト ID をここに設定>'

@app.route('/', methods=["GET", "POST"])
def hello_world():
    if request.method == 'GET':
        post = ""
        image_url = ""
    elif request.method == 'POST':
        try:
            uploaded_file = request.files.get('file')
            if not uploaded_file:
                return 'No file submitted.', 400

            # GCS に画像をアップロード
            client = storage.Client()
            bucket = client.get_bucket(bucket_name)
            file_name = 'image_' + str(int(time()))
            blob = bucket.blob(file_name)
            blob.upload_from_string(
                uploaded_file.read(),
                content_type=uploaded_file.content_type
            )
            image_url = blob.public_url

            # アップロードされた画像のラベルを推定
            client = vision.ImageAnnotatorClient()
            image = vision.Image()
            image.source.image_uri = image_url
            response = client.label_detection(image=image)
            post = Markup('<br>'.join([x.description + ": " + str(x.score) for x in response.label_annotations]))

        except Exception as e:
            post = e.args

    return render_template('index.html', post=post, img=image_url)

if __name__ == "__main__":
    app.run(debug=True)
