from confluent_kafka import Consumer, Producer
import json
from config import KAFKA_CONFIG, KAFKA_TOPICS

class KafkaClient:
    def __init__(self):
        self.consumer = Consumer(KAFKA_CONFIG['consumer'])
        self.producer = Producer(KAFKA_CONFIG['producer'])
        self.consumer.subscribe([KAFKA_TOPICS['requests']])

    def delivery_report(self, err, msg):
        if err is not None:
            print(f'Message delivery failed: {err}')
        else:
            print(f'Message delivered to {msg.topic()}')

    def send_response(self, request_id, data):
        response = {'request_id': request_id}
        response.update(data)

        self.producer.produce(
            KAFKA_TOPICS['results'],
            key=request_id.encode('utf-8') if request_id else None,
            value=json.dumps(response).encode('utf-8'),
            callback=self.delivery_report
        )
        self.producer.flush()

    def send_error(self, error, request_id='unknown'):
        self.send_response(request_id, {'error': str(error)})

    def get_message(self):
        return self.consumer.poll(1.0)

    def close(self):
        self.consumer.close()