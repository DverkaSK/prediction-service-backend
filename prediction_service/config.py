import os
from dotenv import load_dotenv

load_dotenv()  # загружаем переменные из .env

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