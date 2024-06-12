# LB Home Assignment
The implementation is build with the use of https://docs.localstack.cloud/getting-started/installation/#docker-compose - a toolkit for local mocking of AWS services.

## Integration test guide
1. Make sure you have Docker installed.
2. Then you bring the Docker Compose stack up - `docker-compose up --build`. This will run 3 containers: queue consumer, localstack for mocking AWS and aws-cli used as SQS client.
3. In a separate terminal window run `docker-compose exec aws-cli sh` which will get you inside of aws-cli container.
4. Get the URL of created queue:
```
sh-4.2# aws sqs list-queues
{
    "QueueUrls": [
        "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/test-q"
    ]
}
```
5. Add several messages:
```
sh-4.2# aws sqs send-message --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/test-q --message-body '{"
command":"addItem","key":"k1","value":"v1"}'
{
    "MD5OfMessageBody": "7a661e64ffd178b85af90a30338add02",
    "MessageId": "50715dc4-d1ed-45c9-8e72-9d1ba2e1cdc2"
}
sh-4.2# aws sqs send-message --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/test-q --message-body '{"
command":"addItem","key":"k2","value":"v2"}'
{
    "MD5OfMessageBody": "7a661e64ffd178b85af90a30338add02",
    "MessageId": "50715dc4-d1ed-45c9-8e72-9d1ba2e1cdc2"
}
...
```
6. Dump all messages:
```
sh-4.2# aws sqs send-message --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/test-q --message-body '{"command":"getAllItems"}'
{
    "MD5OfMessageBody": "ba8c63e2c9ea366ac83c1cb34e696760",
    "MessageId": "4d74c524-5754-4e94-ba85-efde9d7573b3"
}
```
then you will see the following output of the queue-consumer (in terminal where you ran `docker-compose up`):
```
queue-consumer-1  | 2024/06/12 10:19:57 [INFO] Received 1 messages
queue-consumer-1  | 2024/06/12 10:19:57 [DEBUG] Message: &consumer.commandMsg{Command:"getAllItems", Key:"", Value:""}
queue-consumer-1  | 2024/06/12 10:19:57 [INFO] Receiving messages (wait time 20 sec.)...
queue-consumer-1  | 2024/06/12 10:19:57 [DEBUG] Getting all entries...
queue-consumer-1  | 2024/06/12 10:19:57 [DEBUG] k1 -> v1 @1718187516399816
queue-consumer-1  | 2024/06/12 10:19:57 [DEBUG] k2 -> v2 @1718187590546857
queue-consumer-1  | 2024/06/12 10:19:57 [DEBUG] k3 -> v3 @1718187595612104
```
7. Experiment with other commands - `deleteItem`, `getItem`:
```
sh-4.2# aws sqs send-message --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/test-q --message-body '{"command":"deleteItem","key":"k3"}'
{
    "MD5OfMessageBody": "5f321c21790e82038e2325ba537b4f0c",
    "MessageId": "083dc574-8698-4a9c-b038-4ea7bf2a6031"
}
sh-4.2# aws sqs send-message --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/test-q --message-body '{"command":"getItem","key":"k2"}'
{
    "MD5OfMessageBody": "518aedd349ff07f1227eb6046cbd2b22",
    "MessageId": "e5e1f61f-48ab-4d0f-bd87-eae6763509cf"
}
sh-4.2# 
```
The results of these commands can be also see in the queue-consumer log.

## Some notes and considerations:
 - an ordered map mechanics are achieved using `sync.Map` + `slices.SortFunc` which gives O(1) for insertion, get and removal, and O(nLogn) for `getAllItems` i.e. dump of all items in sorted order. In case of harder CPU/memory limits it'd make sense implementing red-black tree-based map but this would require more time.
 - aws-cli container is used as queue client since in the home assignment there is no much additional logic which would require a separate Go app. (you can also run it in parallel in as many copies as needed)
 - provided `EntryWriter` interface the dump can write data into different locations by implementing e.g. file-based writer
 - in case of distributed setup, the in-memory store should be backed by Redis or similar storage
 - a unit test example can be found under `consumer/store`, run `go test ./...` to run it. 
 - 
