import http.server
import socketserver
import json
import logging
import time
import socket
import base64

# Configure logging
logging.basicConfig(level=logging.DEBUG, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Default IP and port for the emulator (mimics FRC field radio at 10.0.100.5)
HOST = "0.0.0.0"  # Bind to all interfaces
HTTP_PORT = 80  # HTTP port
RADIO_IP = "10.0.100.5"  # Cheesy Arena expects field radio at this IP
AP_PASSWORD = ""  # Set to empty to disable auth; otherwise, use a password

# Simulated field radio state
radio_state = {
    "configured": False,
    "status": "INACTIVE",  # AP status (INACTIVE, CONFIGURING, ACTIVE, ERROR)
    "channel": None,  # Wi-Fi channel
    "stations": {},  # {station_id: {team, ssid, key, vlan}}
    "firmware_version": "1.3.0",  # Mimic VH-109 firmware
    "last_error": None,  # Track last configuration error
    "last_request_time": None  # Track last request timestamp
}

# HTML content for the web interface
WEB_PAGE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FRC Field Radio Emulator</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 font-sans">
    <div class="container mx-auto p-4">
        <h1 class="text-3xl font-bold mb-4 text-center">FRC Field Radio Emulator</h1>
        <div id="radio-info" class="bg-white p-6 rounded-lg shadow-md mb-4">
            <h2 class="text-xl font-semibold mb-2">Field Radio Information</h2>
            <p><strong>Radio IP:</strong> <span id="radio-ip">Loading...</span></p>
            <p><strong>Firmware Version:</strong> <span id="firmware-version">Loading...</span></p>
            <p><strong>Channel:</strong> <span id="channel">None</span></p>
            <p><strong>Status:</strong> <span id="status">INACTIVE</span></p>
            <p><strong>Configured:</strong> <span id="configured-status">Loading...</span></p>
            <p><strong>Last Error:</strong> <span id="last-error">None</span></p>
            <p><strong>Last Request:</strong> <span id="last-request-time">None</span></p>
        </div>
        <div id="config-table" class="bg-white p-6 rounded-lg shadow-md">
            <h2 class="text-xl font-semibold mb-2">Station Configurations</h2>
            <table class="w-full border-collapse">
                <thead>
                    <tr class="bg-gray-200">
                        <th class="border p-2">Station</th>
                        <th class="border p-2">Team</th>
                        <th class="border p-2">SSID</th>
                        <th class="border p-2">WPA Key</th>
                        <th class="border p-2">VLAN ID</th>
                    </tr>
                </thead>
                <tbody id="config-table-body"></tbody>
            </table>
        </div>
    </div>
    <script>
        async function fetchRadioStatus() {
            try {
                const response = await fetch('/status');
                const data = await response.json();
                document.getElementById('radio-ip').textContent = '{RADIO_IP}';
                document.getElementById('firmware-version').textContent = data.firmware_version;
                document.getElementById('channel').textContent = data.channel || 'None';
                document.getElementById('status').textContent = data.status || 'INACTIVE';
                document.getElementById('configured-status').textContent = data.configured ? 'Yes' : 'No';
                document.getElementById('last-error').textContent = data.last_error || 'None';
                document.getElementById('last-request-time').textContent = data.last_request_time || 'None';
                
                const tableBody = document.getElementById('config-table-body');
                tableBody.innerHTML = '';
                for (const [station, config] of Object.entries(data.stations)) {
                    const row = document.createElement('tr');
                    row.innerHTML = `
                        <td class="border p-2">${station}</td>
                        <td class="border p-2">${config.team || 'N/A'}</td>
                        <td class="border p-2">${config.ssid}</td>
                        <td class="border p-2">${config.key || 'None'}</td>
                        <td class="border p-2">${config.vlan}</td>
                    `;
                    tableBody.appendChild(row);
                }
            } catch (error) {
                console.error('Error fetching status:', error);
            }
        }

        // Fetch status on load and every 5 seconds
        window.onload = fetchRadioStatus;
        setInterval(fetchRadioStatus, 5000);
    </script>
</body>
</html>
""".replace('{RADIO_IP}', RADIO_IP)

class FRCFieldRadioEmulator(http.server.BaseHTTPRequestHandler):
    def check_auth(self):
        """Check if the request includes valid Bearer token authentication."""
        if not AP_PASSWORD:
            return True
        auth_header = self.headers.get("Authorization")
        if not auth_header or not auth_header.startswith("Bearer "):
            self.send_response(401)
            self.send_header("WWW-Authenticate", 'Bearer realm="Access Point"')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "error", "message": "Missing or invalid Authorization header"}).encode())
            logger.error("Request failed: Missing or invalid Authorization header")
            return False
        token = auth_header[len("Bearer "):]
        if token != AP_PASSWORD:
            self.send_response(401)
            self.send_header("WWW-Authenticate", 'Bearer realm="Access Point"')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "error", "message": "Invalid Bearer token"}).encode())
            logger.error("Request failed: Invalid Bearer token")
            return False
        return True

    def do_GET(self):
        """Handle GET requests, including the web interface, status, and health endpoints."""
        if not self.check_auth():
            return
        try:
            if self.path == "/":
                self.send_response(200)
                self.send_header("Content-type", "text/html")
                self.end_headers()
                self.wfile.write(WEB_PAGE.encode())
                logger.info("GET / - Served web interface")
            elif self.path == "/status":
                self.send_response(200)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                station_statuses = {
                    station_id: {
                        "ssid": config["ssid"],
                        "hashedWpaKey": "",  # Not implemented
                        "wpaKeySalt": "",  # Not implemented
                        "isLinked": False,  # Emulate no active connections
                        "rxRateMbps": 0.0,
                        "txRateMbps": 0.0,
                        "signalNoiseRatio": 0,
                        "bandwidthUsedMbps": 0.0
                    } for station_id, config in radio_state["stations"].items()
                }
                response = {
                    "status": radio_state["status"],
                    "channel": radio_state["channel"],
                    "stationStatuses": station_statuses,
                    "configured": radio_state["configured"],
                    "firmware_version": radio_state["firmware_version"],
                    "last_error": radio_state["last_error"],
                    "last_request_time": radio_state["last_request_time"]
                }
                response_data = json.dumps(response).encode()
                self.wfile.write(response_data)
                logger.info("GET /status - Returned field radio status")
                logger.debug(f"GET /status - Sent response: {response}")
            elif self.path == "/health":
                self.send_response(200)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                response = {"status": "healthy", "radio_ip": RADIO_IP, "timestamp": time.strftime("%Y/%m/%d %H:%M:%S")}
                self.wfile.write(json.dumps(response).encode())
                logger.info("GET /health - Returned health check")
            else:
                self.send_response(404)
                self.end_headers()
                logger.warning(f"GET {self.path} - Not found")
        except socket.error as e:
            radio_state["last_error"] = f"Socket error during GET: {str(e)}"
            radio_state["status"] = "ERROR"
            logger.error(f"GET {self.path} - Socket error: {str(e)}")

    def do_POST(self):
        """Handle POST requests for field radio configuration."""
        if not self.check_auth():
            return
        if self.path == "/configuration":
            start_time = time.time()
            radio_state["last_request_time"] = time.strftime("%Y/%m/%d %H:%M:%S")
            radio_state["status"] = "CONFIGURING"
            try:
                logger.debug(f"POST /configuration - Headers: {self.headers}")
                content_length = int(self.headers.get("Content-Length", 0))
                post_data = self.rfile.read(content_length).decode()
                logger.info(f"POST /configuration - Received config: {post_data}")
                config = json.loads(post_data)

                # Expect an object with stationConfigurations
                if not isinstance(config, dict) or "stationConfigurations" not in config:
                    radio_state["last_error"] = "Expected an object with stationConfigurations"
                    radio_state["status"] = "ERROR"
                    self.send_response(400)
                    self.send_header("Content-type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps({"status": "error", "message": radio_state["last_error"]}).encode())
                    logger.error("POST /configuration - Expected an object with stationConfigurations")
                    return

                # Update channel (optional)
                radio_state["channel"] = config.get("channel", None)

                # Process station configurations
                stations = config["stationConfigurations"]
                if not isinstance(stations, dict):
                    radio_state["last_error"] = "stationConfigurations must be an object"
                    radio_state["status"] = "ERROR"
                    self.send_response(400)
                    self.send_header("Content-type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps({"status": "error", "message": radio_state["last_error"]}).encode())
                    logger.error("POST /configuration - stationConfigurations must be an object")
                    return

                # Validate and update stations
                radio_state["stations"].clear()  # Reset previous configurations
                for station_id, station_config in stations.items():
                    if not all(key in station_config for key in ["ssid", "wpaKey"]):
                        radio_state["last_error"] = f"Missing ssid or wpaKey for station {station_id}"
                        radio_state["status"] = "ERROR"
                        self.send_response(400)
                        self.send_header("Content-type", "application/json")
                        self.end_headers()
                        self.wfile.write(json.dumps({"status": "error", "message": radio_state["last_error"]}).encode())
                        logger.error(f"POST /configuration - Missing ssid or wpaKey for station {station_id}")
                        return

                    ssid = station_config["ssid"]
                    team_number = ssid  # SSID is team number
                    try:
                        vlan = int(ssid)  # Derive VLAN from SSID
                    except ValueError:
                        radio_state["last_error"] = f"Invalid SSID (must be numeric) for station {station_id}"
                        radio_state["status"] = "ERROR"
                        self.send_response(400)
                        self.send_header("Content-type", "application/json")
                        self.end_headers()
                        self.wfile.write(json.dumps({"status": "error", "message": radio_state["last_error"]}).encode())
                        logger.error(f"POST /configuration - Invalid SSID for station {station_id}")
                        return
                    key = station_config["wpaKey"] if station_config["wpaKey"] else None

                    radio_state["stations"][station_id] = {
                        "team": team_number,
                        "ssid": ssid,
                        "key": key,
                        "vlan": vlan
                    }

                radio_state["configured"] = True
                radio_state["status"] = "ACTIVE"
                radio_state["last_error"] = None

                # Minimal response for FMS
                response = {"status": "success"}
                try:
                    self.send_response(200)
                    self.send_header("Content-type", "application/json")
                    self.end_headers()
                    response_data = json.dumps(response).encode()
                    self.wfile.write(response_data)
                    logger.info(f"POST /configuration - Configured {len(stations)} stations in {time.time() - start_time:.3f}s")
                    logger.debug(f"POST /configuration - Sent response: {response}")
                except socket.error as e:
                    radio_state["last_error"] = f"Failed to send response: {str(e)}"
                    radio_state["status"] = "ERROR"
                    logger.error(f"POST /configuration - Failed to send response: {str(e)}")
            except socket.error as e:
                radio_state["last_error"] = f"Socket error during POST: {str(e)}"
                radio_state["status"] = "ERROR"
                self.send_response(500)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"status": "error", "message": radio_state["last_error"]}).encode())
                logger.error(f"POST /configuration - Socket error: {str(e)}")
            except json.JSONDecodeError:
                radio_state["last_error"] = "Invalid JSON"
                radio_state["status"] = "ERROR"
                self.send_response(400)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"status": "error", "message": "Invalid JSON"}).encode())
                logger.error("POST /configuration - Invalid JSON")
            except Exception as e:
                radio_state["last_error"] = str(e)
                radio_state["status"] = "ERROR"
                self.send_response(500)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"status": "error", "message": str(e)}).encode())
                logger.error(f"POST /configuration - Error: {str(e)}")
        else:
            self.send_response(404)
            self.end_headers()
            logger.warning(f"POST {self.path} - Not found")

def run_server():
    """Start the HTTP server with threading support."""
    http_server_address = (HOST, HTTP_PORT)
    try:
        # Use ThreadingTCPServer for concurrent connections
        socketserver.ThreadingTCPServer.allow_reuse_address = True
        socketserver.ThreadingTCPServer.request_queue_size = 100
        httpd = socketserver.ThreadingTCPServer(http_server_address, FRCFieldRadioEmulator)
        # Set socket options to reduce resets
        httpd.socket.setsockopt(socket.SOL_SOCKET, socket.SO_KEEPALIVE, 1)
        httpd.socket.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
        logger.info(f"Starting FRC Field Radio Emulator on {HOST}:{HTTP_PORT}")
        httpd.serve_forever()
    except Exception as e:
        logger.error(f"Failed to start HTTP server: {str(e)}")
        raise
    finally:
        httpd.server_close()
        logger.info("Shutting down FRC Field Radio Emulator")

if __name__ == "__main__":
    run_server()