![alt text](/assets/icons/icon.png "Logo")

# VBoxSsh
VBoxSsh is a graphical platform independent front end for managing and administrating VirtualBox instances in your network.  
VBoxSsh uses only SSH connections and VBoxManage - no need for a Webserver, additional services and so on.
Yes, I created it due to issues with phpVirtualBox ;-) which requires a webserver, PHP and so on ...
You can also manage your lokal VirtualBox instance.  

VBoxSsh can only do what VBoxManage can do and display.  
For getting the actual status of the virtual machines polling is used.  

VBoxSsh is written in [Go](https://go.dev/) and uses [Fyne](https://fyne.io/) as graphical toolkit.

Author: Reiner Pröls  
Licence: MIT  

## Usage of VBoxSsh
First you have to set a master password. Store it in your brain!  
Then you have to add one or multiple servers and the data needed for  
accesing the vbox account via SSH.  
Then you can control your machines with the buttons un top.  
You you can view informations and change them.  
Most changes are transferred after pressing the "Apply" button.  
I support the most settings you find in the VirtulBox official frontend.  
You can manage storages, create vdi files. Attach USB devices and add ISO images.  
You can export and import ova files.  
You can delete and create virtual machines.  
And you can take and manage snapshots.  

### Adding a server
You can manage your locally installed VirtualBox.  
For this in the SSH fields only give a name for the connection. Leave the other fields empty.  
Virtual machines on local will be started with window. Machines over SSH will be always started headless.  
For a SSH connection enter host and port (normally 22). And give user and password.  
I recommend to use authentication via a key file.  
The passowrd field will be the password for the key - if required. If not leave the password field empty.  
As user you have to choose the account on which VirtualBox runs e.g. vbox in my case.  


## Screenshots
![alt text](/screenshots/main.jpg "Main screen")
![alt text](/screenshots/ssh.jpg "SSH screen")
![alt text](/screenshots/info.jpg "Info screen")
![alt text](/screenshots/system.jpg "System screen")
![alt text](/screenshots/cpu.jpg "CPU/RAM screen")
![alt text](/screenshots/display.jpg "Display screen")
![alt text](/screenshots/rdp.jpg "RDP screen")
![alt text](/screenshots/audio.jpg "Audio screen")
![alt text](/screenshots/storage.jpg "Storage screen")
![alt text](/screenshots/usb.jpg "USB screen")
![alt text](/screenshots/snapshot.jpg "Snapshot screen")
![alt text](/screenshots/task.jpg "Task screen")
![alt text](/screenshots/vmserver.jpg "VM info screen")
![alt text](/screenshots/newhdd.jpg "New HDD info screen")

### Precompiled binaries
#### Linux (64 Bit)
[Tar file](https://github.com/bytemystery-com/vboxssh/releases/download/v0.2.4/VBoxSsh.tar.xz)  
[Standalone binary](https://github.com/bytemystery-com/vboxssh/releases/download/v0.2.4/vboxssh)  
#### Windows (64 Bit)
[Standalone exe](https://github.com/bytemystery-com/vboxssh/releases/download/v0.2.4/VBoxSsh.exe)  
#### Mac
Not available - it could be build but requires Mac + SDK.
#### Android 
[APK](https://github.com/bytemystery-com/vboxssh/releases/download/v0.2.4/VBoxSsh.apk)  
(tablet in landscape mode is recommended)

## Q & A
Q: Where are the server data stored ?  
>A: On Linux it will be located at  
~/.config/fyne/com.bytemystery.vboxssh/preferences.json
On Windows they are under  
C:\Users\<USERNAME>>\AppData\Roaming\fyne\com.bytemystery.vboxssh\preferences.json

Q: Where are the passwords for SSH access are stored ?  
>A: They are encrypted stored in the preferences.json file.  

Q: I have forgotten my master password. How to recover ?
>A: Recovering is not possible.  
Quit the application, delete the preferences.json file and restart application.  
Now you can set a new master password, but I need to give the data for the servers again.

## Compatibility
I have tested with VirtualBox 7.2.4 and with 6.1.50.  
On older VirtualBox some features as secure boot are not available.  

## Statistics
The project with the 2 extra modules I developped for it  
([PicButton](https://github.com/bytemystery-com/picbutton) and [ColorLabel](https://github.com/bytemystery-com/colorlabel))  
consists of round about 15000 lines of go code.


## Links
[Readme](https://bytemystery-com.github.io/vboxssh/)  
[Repository](https://github.com/bytemystery-com/vboxssh/)  
[Issues](https://github.com/bytemystery-com/vboxssh/issues)  
[Discussions](https://github.com/bytemystery-com/vboxssh/discussions/new)  

© Copyright Reiner Pröls, 2026

