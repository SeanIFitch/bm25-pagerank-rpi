import torch
from torch.utils.data import Dataset, DataLoader
import pandas as pd
from sklearn.preprocessing import RobustScaler


class FeatureDataset(Dataset):
    def __init__(self, features, labels, augment=False):
        self.features = torch.tensor(features, dtype=torch.float32)
        self.labels = torch.tensor(labels, dtype=torch.long)
        self.augment = augment

    def __len__(self):
        return len(self.labels)

    def __getitem__(self, idx):
        features = self.features[idx]
        label = self.labels[idx]

        if self.augment:
            # Add small Gaussian noise
            noise = torch.normal(0, 0.01, size=features.shape)
            features = features + noise

        return features, label


def load_and_scale_data(train_file):
    # Load data
    train_df = pd.read_csv(train_file)

    # Separate features and labels
    X_train = train_df.iloc[:, :-1].values
    Y_train = train_df.iloc[:, -1].values

    # Normalize labels to 0 and 1
    Y_train = (Y_train + 1) // 2

    # Scale the features using RobustScaler
    scaler = RobustScaler()
    X_train_scaled = scaler.fit_transform(X_train)

    # Create datasets and dataloaders
    train_dataset = FeatureDataset(X_train_scaled, Y_train, augment=False)

    train_loader = DataLoader(train_dataset, batch_size=16384, shuffle=True, num_workers=4)

    return train_loader, scaler


def load_test_data(test_file, scaler):
    # Load test data
    test_df = pd.read_csv(test_file)
    X_test = test_df.iloc[:, :-1].values
    Y_test = test_df.iloc[:, -1].values

    # Normalize labels to 0 and 1
    Y_test = (Y_test + 1) // 2

    # Scale the features using the pre-fitted scaler
    X_test_scaled = scaler.transform(X_test)

    # Create test dataset and loader
    test_dataset = FeatureDataset(X_test_scaled, Y_test)
    test_loader = DataLoader(test_dataset, batch_size=16384, shuffle=False)

    return test_loader
