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

## Precompiled binaries
In the dis folder are precompiled binaries for Windows (64-Bit) and Linux (64-Bit).  
Apk file for Android will come in future.  
For Mac: I can not compile because MacOS is needed. - But if someone has a Mac you can try it.


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

## Screenshots
![alt text](/example/screenshots/01.png "Screenshot 01")
![alt text](/example/screenshots/02.png "Screenshot 02")
![alt text](/example/screenshots/03.png "Screenshot 03")
![alt text](/example/screenshots/04.png "Screenshot 04")
![alt text](/example/screenshots/05.png "Screenshot 05")

## Q & A
Q: Where are the server data stored ?  
>A: On Linux it will be located at  
~/.config/fyne/com.bytemystery.vboxssh/preferences.json
On Windows they are under  
C:\Users\<USERNAME>>\AppData\Roaming\fyne\com.bytemystery.vboxssh\preferences.json

Q: Where are the passwords for SSH access are stored ?  
>A: They are encrypted stored in the preferences.json file.  

Q: I hav eforgotten my master password. How to recover ?
>A: Recovering is not possible.  
Quit the application, delete the preferences.json file and restart application.  
Now you can set a new master password, but I need to give the data for the servers again.

## Compatibility
I have tested with VirtualBox 7.2.4 and with 6.1.50.  
On older VirtualBox some features as secure boot are not available.  

