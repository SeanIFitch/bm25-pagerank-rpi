import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import Dataset, DataLoader
from torch.optim.lr_scheduler import ReduceLROnPlateau
import pandas as pd
import numpy as np
from sklearn.preprocessing import RobustScaler
from sklearn.feature_selection import SelectKBest, f_classif


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
            nn.BatchNorm1d(out_features),
            nn.LeakyReLU(0.1),
            nn.Dropout(0.3),
            nn.Linear(out_features, out_features),
            nn.BatchNorm1d(out_features)
        )
        self.shortcut = nn.Linear(in_features, out_features) if in_features != out_features else nn.Identity()
        self.activation = nn.LeakyReLU(0.1)

    def forward(self, x):
        return self.activation(self.block(x) + self.shortcut(x))


class ImprovedNeuralNetwork(nn.Module):
    def __init__(self, input_dim):
        super(ImprovedNeuralNetwork, self).__init__()
        self.network = nn.Sequential(
            ResidualBlock(input_dim, 128),
            ResidualBlock(128, 64),
            ResidualBlock(64, 32),
            ResidualBlock(32, 16),
            nn.Linear(16, 2)
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


def train_model(model, train_loader, val_loader, criterion, optimizer, device, epochs=200):
    print(f"Training the model for {epochs} epochs...")

    # Learning rate scheduler
    scheduler = ReduceLROnPlateau(optimizer, mode='min', factor=0.5, patience=10, verbose=True)

    best_val_loss = float('inf')
    early_stop_counter = 0

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

        # Learning rate scheduling and early stopping
        scheduler.step(avg_val_loss)

        if avg_val_loss < best_val_loss:
            best_val_loss = avg_val_loss
            early_stop_counter = 0
            # Optionally save the best model
            torch.save(model.state_dict(), 'best_model.pth')
        else:
            early_stop_counter += 1

        # Early stopping
        if early_stop_counter >= 20:
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
    from sklearn.metrics import (
        precision_score,
        recall_score,
        f1_score,
        confusion_matrix,
        classification_report
    )

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
    train_file = '/home/fitchs/Projects/rpi-search-ranking/data/processed/MSLR-WEB30K/Fold1/1mil-train.csv'
    test_file = '/home/fitchs/Projects/rpi-search-ranking/data/processed/MSLR-WEB30K/Fold1/100k-test.csv'

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
    model = ImprovedNeuralNetwork(input_dim).to(device)

    # Loss and optimizer
    criterion = nn.CrossEntropyLoss(label_smoothing=0.1)  # Added label smoothing
    optimizer = optim.AdamW(model.parameters(), lr=0.001, weight_decay=1e-5)  # Added weight decay for regularization

    # Train the model
    train_model(model, train_loader, val_loader, criterion, optimizer, device)

    # Load the best model
    model.load_state_dict(torch.load('best_model.pth'))

    # Evaluate the model
    evaluate_model(model, test_loader, device)

    # Save the model and scaler
    torch.save(model.state_dict(), 'nn_classifier_model.pth')
    import joblib
    joblib.dump(scaler, 'feature_scaler.joblib')
    print("Model and scaler saved successfully!")


if __name__ == '__main__':
    main()
