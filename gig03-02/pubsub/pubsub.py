import sys, os, time, datetime, json
from google.cloud import pubsub_v1


project_id, topic_name = sys.argv[1], sys.argv[2]
publisher = pubsub_v1.PublisherClient()

# LocalなどService accountのクレデンシャルが必要な場合は以下を利用
# cred_fileにはcredential fileのfile pathを入力

# cred_file = sys.argv[3] 
# publisher = pubsub_v1.publisher.Client.from_service_account_file(cred_file)

topic_path = publisher.topic_path(project_id, topic_name)

while True:
    ymd = datetime.datetime.now().isoformat(" ")
    data = json.dumps({"message":"Hello", "timestamp": ymd})
    data = data.encode("utf-8")
    print("Publish: " + data.decode("utf-8", "ignore") )
    future = publisher.publish(topic_path, data=data)
    print("return ", future.result())
    time.sleep(2)