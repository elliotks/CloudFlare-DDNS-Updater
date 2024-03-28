# Cloudflare DDNS Docker image

This repository contains a Docker image for a Cloudflare Dynamic DNS (DDNS) updater. The Docker image is designed to automatically update your Cloudflare DNS records with your server's current IP address. This is particularly useful if your server is on a dynamic IP address that changes frequently. The repository provides detailed instructions on how to generate a Cloudflare API token, set up the environment configuration, and build and deploy the Docker container.

## Features

- Automatic IP updates: The Docker container automatically updates your Cloudflare DNS records with your server's current IP address.
- Supports both IPv4 and IPv6: You can choose to update your DNS records with either your IPv4 or IPv6 address.
- Customizable update interval: You can set how often the DNS records should be checked and updated.
- Customizable IP Retrieval URLs: The Docker container can be configured to use custom URLs for retrieving your public IPv4 and IPv6 addresses.
- Customizable Time to Live (TTL) value for the updated DNS record in seconds. This determines how long DNS resolvers, which are responsible for directing web traffic, are allowed to cache the DNS record information before checking for updates. A lower TTL value means that changes to your DNS record propagate more quickly, but it also means that DNS resolvers have to make more frequent requests, which can increase load on the DNS server.
- Auto-create DNS records: If the DNS record doesn't exist, the container can automatically create it for you.
- Secure: Uses Cloudflare API tokens for secure access to your Cloudflare account.
- Dockerized: The application is containerized using Docker, making it easy to deploy and run on any system that supports Docker.

## CloudFlare API Token Creation

1. Create a Free Cloudflare Account: If you haven't already, sign up for a free Cloudflare account and add your domain. Cloudflare will guide you through changing your domain's nameservers to Cloudflare's.
2. Get Your Zone ID and API Key:
    1. Log in to your Cloudflare account and navigate to the dashboard.
    2. Click on your profile avatar in the top-right corner and select "My Profile" from the dropdown menu.
    3. In the "API Tokens" section, click on the "Create Token" button.
    4. On the "Create Token" page, select the "Create Custom Token" option.
    5. Give your token a descriptive name, such as "DNS Updater Token" or something similar that helps you identify its purpose.
    6. Under the "Permissions" section, configure the token's permissions as follows:
        - Zone:
          - Zone Settings: Read
          - DNS: Edit
        - Account:
          - Account Settings: Read
    7. In the "Zone Resources" section, select the specific domain(s) you want to grant access to for updating DNS records. You can choose "All zones" if you want the token to have access to all your domains, or select "Specific zone" and enter the domain name(s) you want to update.
    8. In the "Client IP Address Filtering" section, you can optionally specify IP addresses or ranges that are allowed to use this token. If left blank, the token can be used from any IP address.
    9. Set the token's expiration time in the "Token Expiration" section. Choose an appropriate expiration duration based on your security preferences. You can also select "Never" if you want the token to have no expiration.
    10. Click on the "Continue to summary" button to review your token's configuration.
    11. If everything looks correct, click on the "Create Token" button to generate the API token.
    12. On the next page, you will see your newly created API token. Make sure to copy the token value and store it securely, as you won't be able to view it again once you navigate away from this page.
    13. Update your `.env` file with the generated API token, replacing the placeholder value for `CLOUDFLARE_API_TOKEN`.
    - Remember to keep your API token confidential and avoid sharing it with others. If you suspect that your token has been compromised, you can revoke it from the Cloudflare dashboard and create a new one.
    - By following these steps, you can create an API token with the necessary permissions to update DNS records for your desired domain(s) in Cloudflare.

## Setting up the Environment Configuration

There are two options for setting up the environment configuration for the Docker container:

**Option 1: Use an environment file (`stack.env`) with a separate Docker Compose file (`docker-compose-env.yml`).**

This method allows you to keep your environment variables in a separate file for better organization and security. Follow these steps:

1. Copy `stack.env.example` to `stack.env`.
2. Update the `stack.env` file with your specific configuration values.
3. Use the `docker-compose-env.yml` file to deploy the container, which will automatically use the `stack.env` file for environment variables.

**Option 2: Directly set the environment variables inside the `docker-compose.yml` file.**

This method is more straightforward but mixes your configuration with the Docker Compose setup. You can directly add your environment variables in the `docker-compose.yml` file under the `environment` section.

Regardless of the method you choose, you need to update the following configuration items:

**Required Configuration Items:**

- `CLOUDFLARE_API_TOKEN`: This is a token you generate in the Cloudflare dashboard under "My Profile" > "API Tokens". It provides authenticated access to the API without exposing your main account password. Example: `123456789abcdef123456789abcdef12345678`.
- `CLOUDFLARE_DNS_NAME`: The fully qualified domain name (FQDN) that you're updating with your dynamic IP. This should match one of the DNS records in your Cloudflare account. Example: `home.example.com`.

**Optional Configuration Items:**

- `CLOUDFLARE_IP_VERSION`: The IP version to use for updating the DNS record. Valid values: `ipv4`, `ipv6`. Default value: `ipv4`.
- `CLOUDFLARE_IPV4_URL`: The URL to query for retrieving the public IPv4 address. Default value: `https://api.ipify.org`.
- `CLOUDFLARE_IPV6_URL`: The URL to query for retrieving the public IPv6 address. Default value: `https://api64.ipify.org`.
- `CLOUDFLARE_DNS_UPDATE_INTERVAL`: How often to update the Cloudflare DNS record in seconds. The minimum allowed interval is 60 seconds. Default value: `600` (10 minutes).
- `CLOUDFLARE_DNS_TTL`: The Time to Live (TTL) value for the updated DNS record in seconds. Default value: `120` (2 minutes).
- `CLOUDFLARE_AUTO_CREATE_DNS`: Automatically create the DNS record if it doesn't exist. Valid values: `true`, `false`. Default value: `false`.

## Docker Compose Build and Deployment

This project contains two Docker Compose files: `docker-compose.yml` and `docker-compose-env.yml`. The former allows you to directly set environment variables inside the file, while the latter uses an environment file (`stack.env`) for better organization and security.

To build and deploy the Docker container using Docker Compose, follow these steps:

1. Navigate to the directory containing the Docker Compose file you want to use.
2. Build the Docker image by running the following command in your terminal:

    ```bash
    docker compose build
    ```

3. Once the build process is complete, you can start the container with the following command:

    ```bash
    docker compose up -d
    ```

    - The `-d` flag is used to run the container in detached mode, meaning it will run in the background.

4. To check the status of your containers, use the following command:

    ```bash
    docker compose ps
    ```

5. If you need to stop the container, you can do so with the following command:

    ```bash
    docker compose down
    ```

**Note**: Remember to replace any placeholder values in the `docker-compose.yml` file with your actual data before starting the container.
