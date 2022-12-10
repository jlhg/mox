# mox - Comics downloader for Mox.moe

## Installation

```
git clone https://github.com/jlhg/mox.git
cd mox
make
```

## Usage

Create config:

```
cp config/config.example.toml config/config.toml
```

Open `config/config.toml` and Set mox's account E-mail and password.

Download comics books:

```
bin/mox -c config/config.toml dl -i <id>
```

```
id: The comics ID on website's URL
```
