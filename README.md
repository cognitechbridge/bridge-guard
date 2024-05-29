# BridgeGuard: Secure and Encrypted File Sharing System

Welcome to **BridgeGuard**, a secure and encrypted file sharing system designed to enable flexible and secure file sharing with the option to use or bypass traditional cloud ecosystems.

## Table of Contents

- [Introduction](#introduction)
- [How It Works](#how-it-works)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Introduction

**BridgeGuard** is an innovative solution for secure and encrypted file sharing. It allows users to share files securely with the flexibility to use or avoid traditional cloud services such as Google Cloud. The system ensures data privacy and integrity through encryption, providing a robust alternative to centralized cloud storage solutions.

## How It Works

1. **Create a Shared Folder**: Users create a shared folder in an environment of their choice, such as Google Cloud, NAS (Network Attached Storage), or even a removable stick drive.
2. **Mount the Folder**: Use **BridgeGuard** to mount the shared folder to the operating system (OS).
3. **Normal Use**: Once mounted, the drive can be used normally to store and access files.
4. **Encryption**: Underlying operations in **BridgeGuard** ensure that data is encrypted before being stored in the shared folder.
5. **Shared Access**: Other authorized users can open and access the encrypted data using their own encryption keys.

## Features

- **End-to-End Encryption**: Ensures that only authorized users can access shared files.
- **Flexible Storage Options**: Can be used with cloud ecosystems like Google Cloud, NAS, or removable drives.
- **Easy to Use**: Simple commands and configuration for seamless file sharing.
- **Secure Sharing**: Encrypted links for sharing files securely with intended recipients.
- **Shared Environment Independence**: Allows flexibility in choosing or avoiding traditional cloud services.

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) (version 1.16 or later)
- Git

### Steps

1. Clone the repository:
    ```bash
    git clone https://github.com/cognitechbridge/BridgeGuard.git
    cd BridgeGuard
    ```

2. Build the project:
    ```bash
    go build -o bridgeguard
    ```

3. Verify the installation:
    ```bash
    ./bridgeguard --version
    ```

## Usage

### Basic Commands

- **Initialize**: Set up the client for the first time.
    ```bash
    ./bridgeguard init
    ```

- **Mount**: Mount a shared folder to the OS.
    ```bash
    ./bridgeguard mount <shared_folder_path> <mount_point>
    ```
- Use the mounted drive normally to store and access files.
  
## Contributing

We welcome contributions to **BridgeGuard**! If you would like to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Make your changes and commit them (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-branch`).
5. Create a new Pull Request.

Please ensure your code adheres to our coding standards and includes appropriate tests.

## License

This project is licensed under the CC BY-NC 4.0 License. To view a copy of this license, visit [https://creativecommons.org/licenses/by-nc/4.0/](https://creativecommons.org/licenses/by-nc/4.0/).

---

For more information, please visit our [GitHub page](https://github.com/cognitechbridge/BridgeGuard).

Happy Sharing!
