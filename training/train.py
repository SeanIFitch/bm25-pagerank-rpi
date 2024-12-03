import torch
import torch.nn as nn
import torch.optim as optim
import numpy as np
import joblib
from load_data import load_test_data, load_and_scale_data
import nn_model


def train_model(model, train_loader, val_loader, device, epochs=50):
    print(f"Training the model for {epochs} epochs...")

    # Loss and optimizer
    criterion = nn.CrossEntropyLoss(label_smoothing=0.1)
    optimizer = optim.AdamW(
        model.parameters(),
        lr=0.001,
        weight_decay=2e-5
    )

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


def main():
    # Set random seed for reproducibility
    torch.manual_seed(42)
    np.random.seed(42)

    # Device configuration
    device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
    print(f'Using device: {device}')

    # Load and preprocess data
    train_file = '/home/fitchs/Projects/rpi-search-ranking/data/processed/MSLR-WEB30K/Fold1/9.3mil-train.csv'
    val_file = '/home/fitchs/Projects/rpi-search-ranking/data/processed/MSLR-WEB30K/Fold1/1.86mil-vali.csv'

    train, scaler = load_and_scale_data(train_file)
    val = load_test_data(val_file, scaler)

    # Initialize model
    input_dim = train.dataset[0][0].shape[0]
    model = nn_model.NeuralNetwork(input_dim).to(device)

    # Train the model
    train_model(model, train, val, device)

    # Load the best model
    model.load_state_dict(torch.load('../data/models/best_model.pth', weights_only=True))

    # Save the model and scaler
    torch.save(model.state_dict(), '../data/models/nn_classifier_model.pth')
    joblib.dump(scaler, '../data/models/feature_scaler.joblib')
    print("Model and scaler saved successfully!")


if __name__ == '__main__':
    main()
