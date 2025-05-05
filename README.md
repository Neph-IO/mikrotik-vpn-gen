# Mikrotik VPN Generator

A small Go tool to generate OpenVPN client configurations from a MikroTik router.

##  Features

- Generates a TLS certificate and PPP secret
- Downloads the required files from the MikroTik
- Builds a `.ovpn` file from a customizable template
- Packages everything into a ready-to-use `.zip`

##  Requirements

- Go (v1.18+)
- Nginx (to serve the `.zip` files and host the frontend)

##  How to Use

1. Clone the repository:
    ```bash
    git clone https://github.com/neph-IO/mikrotik-vpn-gen.git
    cd mikrotik-vpn-gen
    cp config.example.yaml config.yaml
    ```

2. Edit `config.yaml` to match your MikroTik and environment settings.

3. Edit the `template/Client.ovpn` file to fit your OpenVPN setup.

4. Place your CA certificate (e.g. `ca2025.crt`) in the `template/` folder.

5. Copy the frontend HTML (`frontend/vpncreator.html`) into your Nginx `www/` directory.

6. Run the backend API:
    ```bash
    go run cmd/mikrotik-vpn-gen/init.go
    ```

##  Nginx Configuration

In your Nginx site config, add a reverse proxy to forward API requests:

```nginx
location /api/ {
    proxy_pass http://localhost:8081;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
 ```
## 📝 Notes

- **You must update the frontend's `<select>` dropdown to match your MikroTik profiles.**  

Copy this line (247) as many times as needed:
 ```
<option>OpenVpnProfile1</option>
```
