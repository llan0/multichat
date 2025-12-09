# multichat

A real time Twitch and Kick multichat client (This is an EARLY STAGE prototype)

<img 
  src="https://github.com/user-attachments/assets/95b07338-8458-4aa9-9132-30899b81f406" 
  width="200" 
/>

## How It Works

The app uses a **fan in pattern** to merge multiple chat streams:

```mermaid
graph LR
    TP[Twitch Producer] -->|channel| M[Merge Service]
    KP[Kick Producer] -->|channel| M
    M -->|merged channel| UI[UI Consumer]
```

1. **Producers** (`internal/adapters/`) - Independent producers run in separate goroutines, each streaming messages on their own channel
2. **Service Layer** (`internal/service/`) - Combines all producer channels into a single unified channel
3. **UI** (`internal/ui/`) - Consumes from the merged channel and displays messages in a thread safe Fyne window

## Quick Start

```bash
make run
```

Or

```bash
make build
./multichat
```
