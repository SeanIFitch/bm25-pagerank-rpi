import torch
import joblib
from sklearn.metrics import confusion_matrix, classification_report
from load_data import load_test_data
import nn_model


def test_model(model, test_loader, device):
    print("Testing the model...")
    model.eval()
    correct = 0
    total = 0
    all_predictions = []
    all_labels = []

    with torch.no_grad():
        for features, labels in test_loader:
            features, labels = features.to(device), labels.to(device)
            outputs = model(features)
            _, predicted = torch.max(outputs.data, 1)

            total += labels.size(0)
            correct += (predicted == labels).sum().item()

            # Collect predictions for detailed metrics
            all_predictions.extend(predicted.cpu().numpy())
            all_labels.extend(labels.cpu().numpy())

    accuracy = 100 * correct / total
    print(f'Accuracy on the test set: {accuracy:.2f}%')

    # Additional Metrics
    print("\nClassification Report:")
    print(classification_report(all_labels, all_predictions))

    # Confusion Matrix
    print("\nConfusion Matrix:")
    print(confusion_matrix(all_labels, all_predictions))


def main():
    # Device configuration
    device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
    print(f'Using device: {device}')

    # File paths
    test_file = '/home/fitchs/Projects/rpi-search-ranking/data/processed/MSLR-WEB30K/Fold1/1.86mil-test.csv'

    # Load scaler
    scaler = joblib.load('../data/models/feature_scaler.joblib')

    # Load test data
    test_loader = load_test_data(test_file, scaler)

    # Initialize model
    input_dim = test_loader.dataset[0][0].shape[0]
    model = nn_model.NeuralNetwork(input_dim).to(device)

    # Load model weights
    model.load_state_dict(torch.load('../data/models/nn_classifier_model.pth', map_location=device))

    # Test the model
    test_model(model, test_loader, device)


if __name__ == '__main__':
    main()