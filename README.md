<div align="center">
  <img width="130" alt="icon" src="https://github.com/user-attachments/assets/34e1e11f-8fe4-4689-8340-1a905db13ea9" />
  <h3>multichat</h3>
  <p>A Twitch and Kick multichat client (WIP)</p>
</div>

## How It Works

**Fan in pattern** to merge multiple chat streams:

```mermaid
graph LR
    TP[Twitch Producer] -->|channel| M[Merge Service]
    KP[Kick Producer] -->|channel| M
    M -->|merged channel| UI[UI Consumer]
```

- **Producers** (`internal/adapters/`) - producers run in separate goroutines, each streaming messages on their own channel 
- **Service Layer** (`internal/service/`) - combines all producer channels into a single unified channel 
- **UI** (`internal/ui/`) - consumes from the merged channel and displays messages in a Fyne window 

## Dependencies

- `fyne.io/fyne/v2`
- `github.com/gempir/go-twitch-irc/v4`
- `github.com/coder/websocket`
- `go.uber.org/zap`

Install dependencies:
```bash
go mod download
```

## Running locally 
```bash
git clone git@github.com:llan0/multichat.git
cd multichat
go mod download
make test 
make run
```

## Upcoming
- Auth/credentials management
- Send messages
- Settings/preferences UI
- Render emotes (BBTV, 7TV, FFZ)
- Multiple channel support
