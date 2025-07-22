# Middleware Example

Este exemplo demonstra o uso avançado da biblioteca de IP através de múltiplos middlewares HTTP.

## Funcionalidades

- Múltiplos middlewares em cadeia
- Logging detalhado de IPs
- Segurança e bloqueio de IPs
- Geolocalização simulada
- Rate limiting por IP
- Servidor HTTP completo

## Middleware Chain

The middleware is applied in the following order:

1. **IP Logging** (outermost) - Logs request details
2. **Security** - Validates and secures requests
3. **Geo-Location** - Adds geographic context
4. **Rate Limiting** - Controls request frequency
5. **Application Handler** (innermost) - Serves the actual content

## Running the Example

```bash
cd examples/middleware
go run main.go
```

The server will start on port 8080 with the following endpoints:

- `GET /` - Home page with detailed IP information
- `GET /api/info` - JSON API endpoint with IP data

## Testing the Middleware

### Basic Request
```bash
curl http://localhost:8080
```

### Test with Proxy Headers
```bash
# Test X-Forwarded-For
curl -H 'X-Forwarded-For: 8.8.8.8, 192.168.1.1' http://localhost:8080

# Test Cloudflare headers
curl -H 'CF-Connecting-IP: 1.1.1.1' http://localhost:8080/api/info

# Test multiple headers
curl -H 'X-Forwarded-For: 203.0.113.45' \
     -H 'X-Real-IP: 203.0.113.45' \
     -H 'CF-Connecting-IP: 203.0.113.45' \
     http://localhost:8080
```

### Test Rate Limiting
```bash
# Make multiple rapid requests to trigger rate limiting
for i in {1..15}; do curl http://localhost:8080/api/info; echo; done
```

### Test with IPv6
```bash
curl -H 'X-Forwarded-For: 2001:db8::1' http://localhost:8080
```

## Middleware Features

### 1. IP Logging Middleware
- Extracts real client IP from requests
- Logs request method, path, status code, duration
- Includes IP type classification
- Shows both real IP and remote address

### 2. Security Middleware
- Blocks requests from suspicious IPs
- Adds standard security headers:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `X-XSS-Protection: 1; mode=block`
- Adds client IP information to response headers

### 3. Geo-Location Middleware
- Simulates IP geo-location lookup
- Adds country and city information to headers
- Handles special cases (private IPs, localhost)
- In production, would integrate with real geo-IP services

### 4. Rate Limiting Middleware
- Implements sliding window rate limiting
- Default: 10 requests per minute per IP
- Returns appropriate HTTP status codes
- Adds rate limit headers to responses:
  - `X-RateLimit-Limit`
  - `X-RateLimit-Window`
  - `X-RateLimit-Remaining`

## Response Headers

The middleware adds several informational headers:

- `X-Real-Client-IP`: The real client IP address
- `X-Client-IP-Type`: Type of IP (public, private, etc.)
- `X-Client-Country`: Simulated country information
- `X-Client-City`: Simulated city information
- `X-RateLimit-*`: Rate limiting information

## Production Considerations

### Security Middleware
- In production, maintain blocked IPs in a database
- Consider using external threat intelligence feeds
- Implement more sophisticated security checks

### Geo-Location Middleware
- Use real geo-IP services like MaxMind, IPinfo, or similar
- Cache geo-location data to reduce API calls
- Handle rate limits and failures gracefully

### Rate Limiting Middleware
- Use Redis or similar for distributed rate limiting
- Implement different limits for different endpoints
- Consider user-based limits vs. IP-based limits
- Add whitelist for trusted IPs

### Logging Middleware
- Use structured logging (JSON format)
- Send logs to centralized logging systems
- Include correlation IDs for request tracing
- Sanitize sensitive information

## Key Learning Points

- How to chain HTTP middleware effectively
- Real-world application of IP extraction
- Security best practices for web applications
- Rate limiting implementation strategies
- Middleware design patterns in Go
