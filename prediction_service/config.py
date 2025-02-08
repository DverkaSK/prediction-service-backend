KAFKA_CONFIG = {
    'consumer': {
        'bootstrap.servers': 'localhost:9092',
        'group.id': 'prediction_service',
        'auto.offset.reset': 'earliest'
    },
    'producer': {
        'bootstrap.servers': 'localhost:9092'
    }
}

KAFKA_TOPICS = {
    'requests': 'prediction-requests',
    'results': 'prediction-results'
}

MODEL_PATH = 'C:/Users/asdasd/Downloads/vit.pt'

CLASSES = [
    'Барокко',
    'Древнерусская архитектура',
    'Классицизм',
    'Модерн',
    'Современный и экспериментальный',
    'Сталинская архитектура',
    'Типовая советская архитектура'
]