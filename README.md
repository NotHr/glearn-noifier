# Glearn Notifier

Glearn notifier is a tool which will periodically fetch assignments and quizzes from the GITAM Learning Management System (LMS) and sends push notifications to your device.

## Features
- [x] Assignment notifications
- [x] Real-time updates
- [x] Push notifications via ntfy.sh
- [x] Configurable check intervals
- [ ] API Support
- [ ] Quiz notifications
- [ ] Cloudflare Workers support (JS Port)

## Installation
### Using Pre-built Binaries (Recommended)
1. Download the latest release from the [releases page](https://github.com/nothr/glearn-notifier/releases)
2. Extract the archive:
   ```bash
   # For Linux/MacOS
   tar xzf glearn-notifier_v1.0.0_linux_amd64.tar.gz
   # For Windows
   # Extract the zip file using your preferred tool
   ```
3. Move the binary to a directory in your PATH (Linux/MacOS):
   ```bash
   sudo mv glearn-notifier /usr/local/bin/
   ```

### Building from Source
Requirements:
- Go 1.21 or higher
```bash
git clone https://github.com/nothr/glearn-notifier.git
cd glearn-notifier
go build
```

## Configuration
1. Create a `config.toml` file in the same directory as the binary:
```toml
[credentials]
username = "your_username"  # Your GITAM username
password = "your_password"  # Your GITAM password

[urls]
base = "https://login.gitam.edu"
glearn = "https://glearn.gitam.edu"

[notification]
ntfy_url = "https://ntfy.sh/your-topic"  # Change 'your-topic' to your preferred notification channel
check_delay = "5m"  # Check interval: 5m = 5 minutes
```

2. Subscribe to notifications:
   - Install the ntfy app ([Android](https://play.google.com/store/apps/details?id=io.heckel.ntfy) / [iOS](https://apps.apple.com/us/app/ntfy/id1625396347))
   - Subscribe to your topic (the one you set in `ntfy_url`)

## Usage
### Running as a Regular Program
```bash
./glearn-notifier
```

### Running as a Service (Linux)
1. Create a systemd service file:
```bash
sudo nano /etc/systemd/system/glearn-notifier.service
```

2. Add the following content:
```ini
[Unit]
Description=Glearn Notification Service
After=network.target

[Service]
ExecStart=/usr/local/bin/glearn-notifier
WorkingDirectory=/etc/glearn-notifier
User=your-username
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

3. Create configuration directory and move files:
```bash
sudo mkdir /etc/glearn-notifier
sudo mv config.toml /etc/glearn-notifier/
```

4. Start the service:
```bash
sudo systemctl enable glearn-notifier
sudo systemctl start glearn-notifier
```

5. Check status:
```bash
sudo systemctl status glearn-notifier
```

## Troubleshooting
1. Check logs:
```bash
# If running as a service
sudo journalctl -u glearn-notifier -f
# If running manually
./glearn-notifier 2>&1 | tee glearn.log
```

2. Common issues:
   - **Login Failed**: Check your credentials in config.toml
   - **Network Error**: Verify your internet connection
   - **No Notifications**: Make sure you're subscribed to the correct ntfy topic

## Development
1. Clone the repository:
```bash
git clone https://github.com/yourusername/glearn-notifier.git
```

2. Install dependencies:
```bash
go mod download
```

## Contributing
1. Fork the repository
2. Create your feature branch: `git checkout -b feature/AmazingFeature`
3. Commit your changes: `git commit -m 'Add some AmazingFeature'`
4. Push to the branch: `git push origin feature/AmazingFeature`
5. Open a Pull Request

## License
This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.

## Disclaimer
This tool is not officially affiliated with GITAM University. Use it responsibly and in accordance with the university's policies.
