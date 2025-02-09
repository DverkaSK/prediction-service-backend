import os
from dotenv import load_dotenv

load_dotenv()

KAFKA_CONFIG = {
    'consumer': {
        'bootstrap.servers': os.getenv('KAFKA_BOOTSTRAP_SERVERS'),
        'group.id': 'prediction_service',
        'auto.offset.reset': 'earliest'
    },
    'producer': {
        'bootstrap.servers': os.getenv('KAFKA_BOOTSTRAP_SERVERS')
    }
}

KAFKA_TOPICS = {
    'requests': 'prediction-requests',
    'results': 'prediction-results'
}

MINIO_CONFIG = {
    'endpoint': os.getenv('MINIO_ENDPOINT'),
    'access_key': os.getenv('MINIO_ACCESS_KEY'),
    'secret_key': os.getenv('MINIO_SECRET_KEY'),
    'secure': os.getenv('MINIO_SECURE', 'False').lower() == 'true',
    'bucket_name': os.getenv('MINIO_BUCKET_NAME')
}

MODEL_PATH = os.getenv('MODEL_PATH')

CLASSES = [
    'Барокко',
    'Древнерусская архитектура',
    'Классицизм',
    'Модерн',
    'Современный и экспериментальный',
    'Сталинская архитектура',
    'Типовая советская архитектура'
]