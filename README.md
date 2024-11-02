Quasiauto
=========
this is a fork of https://hg.sr.ht/~ser/quasiauto 
use dotoolc to support wayland

[![Build status](https://builds.sr.ht/~ser/quasiauto/.build.yml.svg)](https://builds.sr.ht/~ser/quasiauto/.build.yml?)
[File a bug here](https://todo.sr.ht/~ser/quasiauto)

Quasiauto performs autotyping and interactive autotyping for [kpmenu](https://github.com/AlessioDP/kpmenu). It's meant to do three things:

1. As a trigger, it's designed to be run by a hotkey (set up in your window manager); it finds the title of the focused window and calls the [kpmenu](https://github.com/AlessioDP/kpmenu) server process, passing it the window title.
2. As an actor, it is called by the [kpmenu](https://github.com/AlessioDP/kpmenu) server process with credentials from a Keepass entry matching the window title. It then either auto-types the information, or enters a quasi-automatic (hence the name) interactive mode.
3. Using a patched robotgo, it can fetch the currently active window title, removing the need for `[kpmenu](https://github.com/AlessioDP/kpmenu)` to use `xdotool` on Linux.

Quasiauto started from a ticket to add autotype to [kpmenu](https://github.com/AlessioDP/kpmenu). The quasi-auto mode is an attempt to improve on autotype in a way that minimizes user interaction while making login data entry more robust than autotype, which is fragile in the face of unnecessarily involved Javascript login pages.

Security
--------

Ignoring the dependencies for Xorg interaction (robotgo), t
here is (relatively) little code and should be easy to audit, even for non-programmers. 
In particular, no network calls are made and this program should run when network jailed.


Installation
============

Binaries can be [downloaded here](https://downloads.ser1.net/software/quasiauto/). Quasiauto can be compiled from source with either `go install`:

```
go install ser1.net/quasiauto
```

or by cloning and building yourself:

```
hg clone https://hg.sr.ht/~ser/quasiauto/
cd quasiauto
go build ./cmd/quasiauto
```


Running
=======

There's only one argument: `-ms <millisecs>`, which defines the amount of "lag" between simulated key presses. It is set to default at 50. All other data, it parses from STDIN.

The intended use is as a utility for [kpmenu](https://github.com/AlessioDP/kpmenu). The idea is that you install `quasiauto` and then bind a hotkey to call `kpmenu --autotype`; [kpmenu](https://github.com/AlessioDP/kpmenu) farms out the actual autotyping part to quasiauto. All in all, it would look like this:

1. Install [kpmenu](https://github.com/AlessioDP/kpmenu); configure it as your KeePass client. **Until the patch is merged, you need the PR that adds autotype support.**
2. Configure [kpmenu](https://github.com/AlessioDP/kpmenu) to use quasiauto to identify the active window by adding this to `~/.config/kpmenu/config`:
```
[executable]
CustomAutotypeWindowID="quasiauto -title"
```
3. Bind a key in your window manager to call kpmenu; for example, in i3:
```
bindsym $mod+r exec /home/ser/.local/bin/kpmenu
bindsym $mod+t exec /home/ser/.local/bin/kpmenu --autotype
```

Now you can trigger autotyping by pressing `MOD-t` on a browser (or other) login window. This depends on having the correct window ID and key sequences set up in the KeePass DB, of course.

A way of testing this is using [zenity](https://gitlab.gnome.org/GNOME/zenity). Say you have a [sourcehut](https://sr.ht/) account, and a corresponding entry named `Sourcehut` in your KeePassDB. Run the following:

```
zenity --forms --add-entry "Username" --add-password "Password" --add-entry "TOTP" --title "Log in to Sourcehut"
```

Make sure the `Username` field is selected and press your `MOD-t` key, and if everything is configured correctly you should get autotyped credentials into the Zenity window.

quasiauto can autotype on behalf of any program that sends it data in the correct format.

On the kpmenu side, there are a few options that can make autotype work better for you:

```
--autotypealwaysconfirm           Always confirm autotype, even when there's only 1 selection
```

This forces kpmenu to *always* ask to start autotype, even if a unique match is made. Without this, autotype can start immediately and -- possibly -- before you've released all of the MOD keys you used to launch autotype, which is less than ideal. I hope to find a way to ensure no keys are pressed before starting autotype, but I don't have a way of doing that right now, so this is a good option to use.

```
--autotypeusersel                 Prompt for autotype entry instead of trying to detect by active window title
```

The kpmenu patch runs a program to find the active window. If you want to disable that, and always select the entry yourself, use this option.

```
--customAutotypeTyper string      Custom executable for autotype typer (default "quasiauto")
```

Set this to `quasiauto -ms 100` to change the type delay.

```
--customAutotypeWindowID string   Custom executable for identifying active window for autotype (default "quasiauto -title")
```

quasiauto can identify the active window, but you could also use, e.g. `xdotool getwindowfocus getwindowname`.  Set this option to change how window titles are selected.


Limitations
================

VKEY is not implemented, as it's windows specific and I don't have access to a Windows computer.

I haven't found a way to ensure that no modifier keys are being held when autotype starts. Consequently, there's a risk that whatever hotkey trigger is set up might mean a meta key is still being held when typing starts. This is Not Good™, and avoiding this means you may have to insert a small delay (500ms) before executing the program. In my case, the various delays introduced by forking, parsing input, and so on is enough that I don't have a problem if I don't linger on the modifier; however, YMMV. I'll add a safety measure, if I can find a way to read the keyboard state and ensure it's clear in a platform-independent way.


Development
===========

Discussion of the feature is in a github ticket. The design is meant to 

1. Limit changes to [kpmenu](https://github.com/AlessioDP/kpmenu), and in particular, minimize added dependencies. Autotype requires UI interaction, and any UI library dependency -- and especially any cross-platform one -- has large overhead. Dependencies increase the code that must be reviewed for security, so this is an important consideration for [kpmenu](https://github.com/AlessioDP/kpmenu).
2. Maintain the [kpmenu](https://github.com/AlessioDP/kpmenu) security model, which rests on the fact that clients can not extract secret information from the server process. Secrets are available to only the server process or processes that it forks (e.g., xsel).

Quasiauto reads STDIN and expects (quasi-BNF):


```
sequence  ::= cmds "\n"
cmds      ::= cmds "{" cmd "}" | ""
cmd       ::= key | "TAB" | "ENTER" | "TOTP" | "DELAY" digits
digits    ::= [0-9]+
fields    ::= fields "\n" field | field
field     ::= key "\t" value
key       ::= [^\n\t]+
value     ::= [^n]+
```

For example:

```
{USERNAME}{TAB}{PASSWORD}{TAB}{TOTP}{ENTER}
Username	John B.
Password	BooTAY, BooTAY
OTP		1328765
```

Metrics
-------

```24
         Coverage     Parse benchmark 	Exec benchmark
v0.2.0   94%          12734 		10801
v0.1.0   78%           8035
```
