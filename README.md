gomon
=====

go source file monitor, which restarts/rebuilds your go package automatically
while you are changing it.

What's the difference?
=====

The original version didn't kill the child process correctly and it stops when the child process is in infinite loop. This version can kill the child process correctly even in infinite loop.

Also I removed the most of the options because I don't use them and only specify custom command.

Install
-------

    go get -u github.com/honteng/gomon

Usage
-----

    gomon [dir] -- [cmd]

Monitoring With Custom Command:

    gomon src -- go run -x server.go # execute go run -x server.go
    gomon src -- go build -x package # execute go build -x package

Recursively check the subfolders

    gomon src -R -- go run -x server.go

Only watch the mached files with regex

    gomon src -m '.*go' -- cmd 

Ignore the specific file gomon src -d 'ignore.go' -- cmd 

    gomon src -d 'ignore.go' -- cmd 

Screenshot
----------

![](https://raw.github.com/c9s/gomon/gh-pages/images/screenshot.png)

Todo
-----

- Add configration file support.
- Command queue support.


Related Product
---------------

GoTray <http://gotray.extremedev.org/>


Contributors
------------

- Ask Bjørn Hansen
- Yasuhiro Matsumoto (a.k.a mattn)

License
--------

MIT License



[![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/c9s/gomon/trend.png)](https://bitdeli.com/free "Bitdeli Badge")



[![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/c9s/gomon/trend.png)](https://bitdeli.com/free "Bitdeli Badge")

