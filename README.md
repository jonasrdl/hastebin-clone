# hastebin clone

## Setup

1. Install latest go https://go.dev/doc/install
2. Build a binary `go build -o hastebin-clone`
3. Run the binary with env arguments:
4. 
`API_KEY=x [PORT=x] DB_USER=x DB_PASS=x DB_NAME=x [DB_PORT=x] ./hastebin-clone`   
`PORT` defaults to `8080` if no port is provided   
`DB_PORT` defaults to `3306` if no db port is provided   
`API_KEY` is the API key you want the application to use (this one you have to send
with your post requests)


## Endpoints

### Create a New Paste

Create a new paste with a unique ID and password.

- **Endpoint:** `POST /`
- **Request Payload:**
    - Content-Type: application/json
    - Body: `{ "content": "Your paste content" }`
    - Headers:
      - `Authorization: <api-key>`
- **Success Response (201):**
    - Content-Type: application/json
    - Body: `{ "id": "unique-paste-id", "password": "generated-password" }`
- **Error Response (400):**
    - Content-Type: application/json
    - Body: `{ "error": "Invalid input" }`
    - `{ "error": "Invalid API key" }`

### Get Paste Content

Retrieve the content of a paste using its ID. Requires authentication with the paste's password.

- **Endpoint:** `GET /{id}`
- **Path Parameters:**
    - `{id}`: Paste ID
- **Headers:**
    - `Authorization: Basic {base64-encoded-password}`
- **Query Parameter:**
  - `password=<password>`
- **Success Response (200):**
    - Content-Type: text/plain
    - Body: `Paste content`
- **Error Responses:**
    - Content-Type: application/json
    - 401 Unauthorized: `{ "error": "Invalid password" }`
    - 404 Not Found: `{ "error": "Paste not found" }`