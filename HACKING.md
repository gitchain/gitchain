Hacking notes
=============

It is very convenient to create a couple of config files like these:

```
[general]
data-path=gitchain1.db
[api]
http-port=3001
[network]
port=31001
```

```
[general]
data-path=gitchain2.db
[api]
http-port=3002
[network]
port=31002
join=localhost:31001
```

Lets say you called them git1.config and git2.config, now you can start them as
simply as this:

```
$ gitchain --config git1.config
$ gitchain --config git2.config
```
