import torch
from torchvision import transforms
from PIL import Image
from transformers import ViTForImageClassification
from torch.serialization import add_safe_globals
from config import CLASSES, MODEL_PATH

class StylePredictor:
    def __init__(self):
        add_safe_globals([ViTForImageClassification])
        self.model = torch.load(MODEL_PATH, weights_only=False)
        self.model.eval()

        self.transform = transforms.Compose([
            transforms.Resize((224, 224)),
            transforms.ToTensor(),
            transforms.Normalize(
                mean=[0.485, 0.456, 0.406],
                std=[0.229, 0.224, 0.225]
            )
        ])

    def predict(self, image_path):
        try:
            image = Image.open(image_path).convert('RGB')
            image = self.transform(image).unsqueeze(0)

            with torch.no_grad():
                outputs = self.model(image)
                probabilities = torch.nn.functional.softmax(outputs.logits, dim=1)

            max_prob, predicted = torch.max(probabilities, 1)
            main_style = CLASSES[predicted.item()]
            main_confidence = max_prob.item()

            all_probabilities = [
                {"style": CLASSES[idx], "probability": prob}
                for idx, prob in enumerate(probabilities[0].tolist())
            ]
            all_probabilities.sort(key=lambda x: x["probability"], reverse=True)

            return {
                "main_prediction": {
                    "style": main_style,
                    "confidence": main_confidence
                },
                "all_probabilities": all_probabilities
            }, None
        except Exception as e:
            return None, str(e)