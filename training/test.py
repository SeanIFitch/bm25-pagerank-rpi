import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import Dataset, DataLoader
import pandas as pd
import numpy as np
from sklearn.preprocessing import RobustScaler
from sklearn.feature_selection import SelectKBest, f_classif
import joblib
from sklearn.metrics import (
    confusion_matrix,
    classification_report
)

# Results:
"""
Training the model for 50 epochs...
Epoch [1/50], Train Loss: 0.7401, Val Loss: 0.5928
Epoch [2/50], Train Loss: 0.5914, Val Loss: 0.5806
Epoch [3/50], Train Loss: 0.5808, Val Loss: 0.5717
Epoch [4/50], Train Loss: 0.5728, Val Loss: 0.5667
Epoch [5/50], Train Loss: 0.5654, Val Loss: 0.5596
Epoch [6/50], Train Loss: 0.5597, Val Loss: 0.5558
Epoch [7/50], Train Loss: 0.5558, Val Loss: 0.5505
Epoch [8/50], Train Loss: 0.5542, Val Loss: 0.5487
Epoch [9/50], Train Loss: 0.5475, Val Loss: 0.5419
Epoch [10/50], Train Loss: 0.5431, Val Loss: 0.5462
Epoch [11/50], Train Loss: 0.5392, Val Loss: 0.5360
Epoch [12/50], Train Loss: 0.5356, Val Loss: 0.5454
Epoch [13/50], Train Loss: 0.5321, Val Loss: 0.5258
Epoch [14/50], Train Loss: 0.5291, Val Loss: 0.5268
Epoch [15/50], Train Loss: 0.5261, Val Loss: 0.5349
Epoch [16/50], Train Loss: 0.5230, Val Loss: 0.5225
Epoch [17/50], Train Loss: 0.5675, Val Loss: 0.5451
Epoch [18/50], Train Loss: 0.5869, Val Loss: 0.5347
Epoch [19/50], Train Loss: 0.5385, Val Loss: 0.5320
Epoch [20/50], Train Loss: 0.5364, Val Loss: 0.5326
Epoch [21/50], Train Loss: 0.5319, Val Loss: 0.5284
Epoch [22/50], Train Loss: 0.5279, Val Loss: 0.5270
Epoch [23/50], Train Loss: 0.5279, Val Loss: 0.5216
Epoch [24/50], Train Loss: 0.5178, Val Loss: 0.5126
Epoch [25/50], Train Loss: 0.5119, Val Loss: 0.5125
Epoch [26/50], Train Loss: 0.5082, Val Loss: 0.5072
Epoch [27/50], Train Loss: 0.5082, Val Loss: 0.5111
Epoch [28/50], Train Loss: 0.5065, Val Loss: 0.5058
Epoch [29/50], Train Loss: 0.5004, Val Loss: 0.5013
Epoch [30/50], Train Loss: 0.5036, Val Loss: 0.5021
Epoch [31/50], Train Loss: 0.4974, Val Loss: 0.5039
Epoch [32/50], Train Loss: 0.4943, Val Loss: 0.4979
Epoch [33/50], Train Loss: 0.4920, Val Loss: 0.4997
Epoch [34/50], Train Loss: 0.4897, Val Loss: 0.4948
Epoch [35/50], Train Loss: 0.4876, Val Loss: 0.4962
Epoch [36/50], Train Loss: 0.4852, Val Loss: 0.4951
Epoch [37/50], Train Loss: 0.4832, Val Loss: 0.4928
Epoch [38/50], Train Loss: 0.4813, Val Loss: 0.4954
Epoch [39/50], Train Loss: 0.4795, Val Loss: 0.4929
Epoch [40/50], Train Loss: 0.4777, Val Loss: 0.4918
Epoch [41/50], Train Loss: 0.4760, Val Loss: 0.4898
Epoch [42/50], Train Loss: 0.4802, Val Loss: 0.5037
Epoch [43/50], Train Loss: 0.4857, Val Loss: 0.4924
Epoch [44/50], Train Loss: 0.5136, Val Loss: 0.5465
Epoch [45/50], Train Loss: 0.5503, Val Loss: 0.5113
Epoch [46/50], Train Loss: 0.5149, Val Loss: 0.5051
Epoch [47/50], Train Loss: 0.5173, Val Loss: 0.5042
Epoch [48/50], Train Loss: 0.4950, Val Loss: 0.5009
Epoch [49/50], Train Loss: 0.4909, Val Loss: 0.4984
Epoch [50/50], Train Loss: 0.4841, Val Loss: 0.4924
Evaluating the model on the test set...
Accuracy on test set: 81.92%

Detailed Classification Report:
              precision    recall  f1-score   support

           0       0.82      0.83      0.82    500705
           1       0.82      0.81      0.82    499295

    accuracy                           0.82   1000000
   macro avg       0.82      0.82      0.82   1000000
weighted avg       0.82      0.82      0.82   1000000


Confusion Matrix:
[[413275  87430]
 [ 93363 405932]]
"""

class FeatureDataset(Dataset):
    def __init__(self, features, labels):
        self.features = torch.tensor(features, dtype=torch.float32)
        self.labels = torch.tensor(labels, dtype=torch.long)

    def __len__(self):
        return len(self.labels)

    def __getitem__(self, idx):
        return self.features[idx], self.labels[idx]


class ResidualBlock(nn.Module):
    def __init__(self, in_features, out_features):
        super(ResidualBlock, self).__init__()
        self.block = nn.Sequential(
            nn.Linear(in_features, out_features),
            nn.LayerNorm(out_features),
            nn.GELU(),
            nn.Dropout(0.3),
            nn.Linear(out_features, out_features),

            nn.BatchNorm1d(out_features)
        )
        self.shortcut = nn.Linear(in_features, out_features) if in_features != out_features else nn.Identity()
        self.activation = nn.GELU()

    def forward(self, x):
        return self.activation(self.block(x) + self.shortcut(x))


class NeuralNetwork(nn.Module):
    def __init__(self, input_dim):
        super(NeuralNetwork, self).__init__()
        self.network = nn.Sequential(
            # More residual blocks
            ResidualBlock(input_dim, 128),
            ResidualBlock(128, 256),
            ResidualBlock(256, 512),
            ResidualBlock(512, 256),
            ResidualBlock(256, 128),
            ResidualBlock(128, 64),
            nn.Dropout(0.4),
            nn.Linear(64, 2)
        )

    def forward(self, x):
        return self.network(x)


def load_and_preprocess_data(train_file, test_file):
    # Load data
    train_df = pd.read_csv(train_file)
    test_df = pd.read_csv(test_file)

    # Separate features and labels
    X_train = train_df.iloc[:, :-1].values
    X_test = test_df.iloc[:, :-1].values
    Y_train = train_df.iloc[:, -1].values
    Y_test = test_df.iloc[:, -1].values

    # Normalize labels to 0 and 1
    Y_train = (Y_train + 1) // 2
    Y_test = (Y_test + 1) // 2

    # Feature selection
    selector = SelectKBest(f_classif, k=min(50, X_train.shape[1]))
    X_train_selected = selector.fit_transform(X_train, Y_train)
    X_test_selected = selector.transform(X_test)

    # Scale the features using RobustScaler
    scaler = RobustScaler()
    X_train_scaled = scaler.fit_transform(X_train_selected)
    X_test_scaled = scaler.transform(X_test_selected)

    return X_train_scaled, X_test_scaled, Y_train, Y_test, scaler


def train_model(model, train_loader, val_loader, criterion, optimizer, device, epochs=50):
    print(f"Training the model for {epochs} epochs...")

    # Learning rate scheduler
    scheduler = torch.optim.lr_scheduler.ReduceLROnPlateau(
        optimizer,
        mode='min',
        factor=0.5,
        patience=10,
        min_lr=1e-5
    )

    best_val_loss = float('inf')
    early_stop_counter = 0
    patience = 15

    model.train()
    for epoch in range(epochs):
        total_train_loss = 0
        for batch_features, batch_labels in train_loader:
            batch_features, batch_labels = batch_features.to(device), batch_labels.to(device)

            # Zero the parameter gradients
            optimizer.zero_grad()

            # Forward pass
            outputs = model(batch_features)
            loss = criterion(outputs, batch_labels)

            # Backward pass and optimize
            loss.backward()
            optimizer.step()

            total_train_loss += loss.item()

        # Validation
        model.eval()
        total_val_loss = 0
        with torch.no_grad():
            for val_features, val_labels in val_loader:
                val_features, val_labels = val_features.to(device), val_labels.to(device)
                val_outputs = model(val_features)
                val_loss = criterion(val_outputs, val_labels)
                total_val_loss += val_loss.item()

        avg_train_loss = total_train_loss / len(train_loader)
        avg_val_loss = total_val_loss / len(val_loader)

        print(f'Epoch [{epoch + 1}/{epochs}], Train Loss: {avg_train_loss:.4f}, Val Loss: {avg_val_loss:.4f}')

        # Learning rate scheduling with validation loss
        scheduler.step(avg_val_loss)

        if avg_val_loss < best_val_loss:
            best_val_loss = avg_val_loss
            early_stop_counter = 0
            # Save the best model
            torch.save(model.state_dict(), '../data/models/best_model.pth')
        else:
            early_stop_counter += 1

        # Early stopping
        if early_stop_counter >= patience:
            print(f"Early stopping triggered after {epoch + 1} epochs")
            break

        model.train()


def evaluate_model(model, test_loader, device):
    print("Evaluating the model on the test set...")
    model.eval()
    correct = 0
    total = 0

    # For detailed metrics
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
    print(f'Accuracy on test set: {accuracy:.2f}%')

    # Additional Metrics
    print("\nDetailed Classification Report:")
    print(classification_report(all_labels, all_predictions))

    # Confusion Matrix
    cm = confusion_matrix(all_labels, all_predictions)
    print("\nConfusion Matrix:")
    print(cm)


def main():
    # Set random seed for reproducibility
    torch.manual_seed(42)
    np.random.seed(42)

    # Device configuration
    device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
    print(f'Using device: {device}')

    # Load and preprocess data
    train_file = '/home/fitchs/Projects/rpi-search-ranking/data/processed/MSLR-WEB30K/Fold1/9mil-train.csv'
    test_file = '/home/fitchs/Projects/rpi-search-ranking/data/processed/MSLR-WEB30K/Fold1/1mil-test.csv'

    X_train, X_test, y_train, y_test, scaler = load_and_preprocess_data(train_file, test_file)

    # Split training data into training and validation sets
    from sklearn.model_selection import train_test_split
    X_train, X_val, y_train, y_val = train_test_split(X_train, y_train, test_size=0.2, random_state=42)

    # Create datasets and dataloaders
    train_dataset = FeatureDataset(X_train, y_train)
    val_dataset = FeatureDataset(X_val, y_val)
    test_dataset = FeatureDataset(X_test, y_test)

    train_loader = DataLoader(train_dataset, batch_size=16384, shuffle=True, num_workers=4)
    val_loader = DataLoader(val_dataset, batch_size=16384, shuffle=False, num_workers=4)
    test_loader = DataLoader(test_dataset, batch_size=16384, shuffle=False, num_workers=4)

    # Initialize model
    input_dim = X_train.shape[1]
    model = NeuralNetwork(input_dim).to(device)

    # Loss and optimizer
    criterion = nn.CrossEntropyLoss(label_smoothing=0.1)
    optimizer = optim.AdamW(
        model.parameters(),
        lr=0.001,
        weight_decay=2e-5
    )

    # Train the model
    train_model(model, train_loader, val_loader, criterion, optimizer, device)

    # Load the best model
    model.load_state_dict(torch.load('../data/models/best_model.pth', weights_only=True))

    # Evaluate the model
    evaluate_model(model, test_loader, device)

    # Save the model and scaler
    torch.save(model.state_dict(), '../data/models/nn_classifier_model.pth')
    joblib.dump(scaler, '../data/models/feature_scaler.joblib')
    print("Model and scaler saved successfully!")


if __name__ == '__main__':
    main()
