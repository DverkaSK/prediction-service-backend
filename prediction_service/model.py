import torch
from torchvision import transforms
from PIL import Image
from transformers import ViTForImageClassification
from torch.serialization import add_safe_globals
from config import CLASSES, MODEL_PATH, MINIO_CONFIG
from minio import Minio
import io
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def get_style_names():
    return CLASSES


class StylePredictor:
    def __init__(self):
        try:
            logger.info("Initializing model...")
            add_safe_globals([ViTForImageClassification])
            self.model = torch.load(MODEL_PATH, weights_only=False)
            self.model.eval()
            logger.info("Model loaded successfully")

            self.transform = transforms.Compose([
                transforms.Resize((224, 224)),
                transforms.ToTensor(),
                transforms.Normalize(
                    mean=[0.485, 0.456, 0.406],
                    std=[0.229, 0.224, 0.225]
                )
            ])

            self.minio_client = Minio(
                MINIO_CONFIG['endpoint'],
                access_key=MINIO_CONFIG['access_key'],
                secret_key=MINIO_CONFIG['secret_key'],
                secure=MINIO_CONFIG['secure']
            )

            if not self.minio_client.bucket_exists("images"):
                logger.error("Bucket 'images' does not exist!")
            else:
                logger.info("Successfully connected to MinIO")

        except Exception as e:
            logger.error(f"Error during initialization: {str(e)}")
            raise

    def predict(self, image_id):
        try:
            logger.info(f"Processing image with ID: {image_id}")

            try:
                data = self.minio_client.get_object(MINIO_CONFIG['bucket_name'], image_id)
                image_bytes = io.BytesIO(data.read())
                image = Image.open(image_bytes).convert('RGB')
                logger.info(f"Image loaded successfully. Size: {image.size}")
            except Exception as e:
                logger.error(f"Error loading image from MinIO: {str(e)}")
                return None, f"Failed to load image: {str(e)}"

            image_tensor = self.transform(image).unsqueeze(0)

            with torch.no_grad():
                outputs = self.model(image_tensor)
                probabilities = torch.nn.functional.softmax(outputs.logits, dim=1)

            max_prob, predicted = torch.max(probabilities, 1)
            main_style = CLASSES[predicted.item()]
            main_confidence = max_prob.item()

            all_probabilities = [
                {"style": CLASSES[idx], "probability": float(prob)}
                for idx, prob in enumerate(probabilities[0].tolist())
            ]
            all_probabilities.sort(key=lambda x: x["probability"], reverse=True)

            response = {
                "main_prediction": {
                    "style": main_style,
                    "confidence": main_confidence
                },
                "all_probabilities": all_probabilities
            }

            logger.info("Prediction completed successfully")
            return response, None

        except Exception as e:
            logger.error(f"Error during prediction: {str(e)}")
            return None, str(e)

