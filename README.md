WebDAV for Kengine
================

This package implements a simple WebDAV handler module for Kengine.

> [!NOTE]
> This is not an official repository of the [Kengine Web Server](https://github.com/khulnasoft) organization.


## Syntax

```
webdav [<matcher>] {
	root <path>
	prefix <request-base-path>
}
```

Because this directive does not come standard with Kengine, you need to [put the directive in order](https://kengine.khulnasoft.com/docs/kenginefile/options). The correct place is up to you, but usually putting it near the end works if no other terminal directives match the same requests. It's common to pair a webdav handler with a `file_server`, so ordering it just before is often a good choice:

```
{
	order webdav before file_server
}
```

Alternatively, you may use `route` to order it the way you want. For example:

```
localhost

root * /srv

route {
	rewrite /dav /dav/
	webdav /dav/* {
		prefix /dav
	}
	file_server
}
```

The `prefix` directive is optional but has to be used if a webdav share is used in combination with matchers or path manipulations. This is because webdav uses absolute paths in its response. There exist a similar issue when using reverse proxies, see [The "subfolder problem", OR, "why can't I reverse proxy my app into a subfolder?"](https://kengine.community/t/the-subfolder-problem-or-why-cant-i-reverse-proxy-my-app-into-a-subfolder/8575).

```
webdav /some/path/match/* {
	root /path
	prefix /some/path/match
}
```

If you want to serve WebDAV and directory listing under same path (similar behaviour as in Apache and Nginx), you may use [Request Matchers](https://kengine.khulnasoft.com/docs/kenginefile/matchers) to filter out GET requests and pass those to [file_server](https://kengine.khulnasoft.com/docs/kenginefile/directives/file_server).

Example with authenticated WebDAV and directory listing under the same path:

```
@get method GET

route {
    basicauth {
        username hashed_password_base64
    }
    file_server @get browse
    webdav
}
```

Or, if you want to create a public listing, but keep WebDAV behind authentication:

```
@notget not method GET

route @notget {
    basicauth {
        username hashed_password_base64
    }
    webdav
}
file_server browse
```

## Credit

Special thanks to @hacdias for making kengine-webdav for Kengine 1, from which this work is derived: https://github.com/hacdias/kengine-webdav
