# CoreDNS Containerd Plugin
Experimental CoreDNS plugin to perform lookup and resolution via container labels.

# Usage
Place this in your `corefile`:

```
.:1053 {
  log
  debug
  containerd /run/containerd/containerd.sock default
}
```

Where `/run/containerd/containerd.sock` is the path to the containerd socket and `default` is the desired containerd namespace.
