import json
from model import StylePredictor
from kafka_client import KafkaClient

def process_requests():
    predictor = StylePredictor()
    kafka_client = KafkaClient()

    print("Starting Kafka consumer...")

    try:
        while True:
            msg = kafka_client.get_message()

            if msg is None:
                continue
            if msg.error():
                print(f"Consumer error: {msg.error()}")
                continue

            request_id = 'unknown'
            try:
                request_data = json.loads(msg.value().decode('utf-8'))
                image_path = request_data.get('path')
                request_id = request_data.get('request_id', 'unknown')

                print(f"Processing request {request_id}")

                prediction_result, error = predictor.predict(image_path)

                if error is None:
                    kafka_client.send_response(request_id, prediction_result)
                else:
                    kafka_client.send_error(error, request_id)

                print(f"Sent response for request {request_id}")

            except Exception as e:
                print(f"Error processing message: {e}")
                kafka_client.send_error(str(e), request_id)

    except KeyboardInterrupt:
        print("Shutting down...")
    finally:
        kafka_client.close()

if __name__ == "__main__":
    process_requests()