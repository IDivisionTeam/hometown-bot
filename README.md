# hometown-bot ![hometown-bot](https://img.shields.io/badge/hometown-bot-green.svg)

## Overview

A highly customizable Discord bot for management of temporary channels.

## Features

- Automatic Channel Creation and Deletion: Automatically creates new voice channels when users join and deletes them
  when they are empty.
- Channel Organization: Renames and changes the capacity of the channels based on user-defined settings.
- User Permissions Management: Ensures only authorized users can create or change certain channels.

## Usage

```discord
/command sub-command [arguments]
```

## Commands

There are 3 types of commands available:

- `lobby` - manage and organize voice channels within your Discord server efficiently.
- `reset` - restore default settings for lobbies to maintain consistency.
- `message` - facilitate communication across channels with targeted messaging.

### Lobby

- `register` `<channel>` - Registers a new `lobby` with the specified `channel`. After registration, the lobby is
  available for users to join and interact.

```slash-command
/lobby register <channel>
```

- `capacity` `<lobby>` `<capacity>` - Sets the maximum number of users allowed in the specified `lobby`.

```slash-command
/lobby capacity <lobby> <capacity>
```

- `name` `<lobby>` `<name>` - Defines the default name for new channels created in `lobby`.

```slash-command
/lobby name <lobby> <name>
```

- `list` - Displays a list of all currently registered lobbies. Useful for server administrators to review and manage
  existing channels.

```slash-command
/lobby list
```

- `remove` `<lobby>` - Deletes the specified `lobby` from the server. After removal, the lobby is no longer available
  for users to join and interact.

```slash-command
/lobby remove <lobby>
```

### Reset

- `lobby name` `<lobby>` - Restores the name of `lobby` to its default setting (Кімната 'username').

```slash-command
/reset lobby name <lobby>
```

- `lobby capacity 'lobby'` - Resets the maximum user capacity of `lobby` to its default setting (unlimited).

```slash-command
/reset lobby capacity <lobby>
```

### Message

- `all` `<channel>` `<message>` - Sends a message `message` to the specified `channel`.

```slash-command
/message all <channel> <message>
```

## Examples

```slash-command
/lobby list

Active Lobbies:
Name: Duo, Channel template: Duo %username%, Capacity: 2
```

```slash-command
/lobby register Trio

Lobby "Trio" successfully registered.
```

```slash-command
/reset capacity Duo

Capacity successfully reset for "Duo".

/lobby list

Active Lobbies:
Name: Duo, Channel template: Duo %username%, Capacity: unlimited
```
