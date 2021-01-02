# Ephemeral Server

Hosting as you need it.

Provides a wrapper around `terraform` and `ansible` that will automatically install
a variety of programs onto a digital ocean droplet. Programs are installed onto a
persistent volume that persist after the more expensive VPS is shut down.

## Usage

E.g. boot up vanilla Minecraft server
```
DIGITAL_OCEAN_TOKEN=<insert your token here> ./ephemeralctl.sh -n somename -t vanilla-latest -c
```

```
./ephemeralctl.sh -n somename -t vanilla-latest -c
digitalocean_volume.mc_vol: Creating...
digitalocean_droplet.minecraft: Creating...
digitalocean_volume.mc_vol: Creation complete after 7s [id=c2559230-4cef-11eb-a8e9-0a58ac14549e]
digitalocean_droplet.minecraft: Still creating... [10s elapsed]
digitalocean_droplet.minecraft: Still creating... [20s elapsed]
digitalocean_droplet.minecraft: Creation complete after 23s [id=224454944]
digitalocean_volume_attachment.minecraft-vol-attach: Creating...
digitalocean_volume_attachment.minecraft-vol-attach: Still creating... [10s elapsed]
digitalocean_volume_attachment.minecraft-vol-attach: Creation complete after 12s [id=224454944-c2559230-4cef-11eb-a8e9-0a58ac14549e-20210102114317282200000001]

Apply complete! Resources: 3 added, 0 changed, 0 destroyed.

Outputs:

ip = 159.89.95.238
```

Then run:
```
./ephemeralctl.sh -n somename -t vanilla-latest -I
```

```
PLAY [minecraft] **********************************************************************************

TASK [Gathering Facts] ****************************************************************************
The authenticity of host '159.89.95.238 (159.89.95.238)' can't be established.
ECDSA key fingerprint is SHA256:2kNf/Eo3vYuEevW8nGq7j7odx/MTrJAumMlV1WBcmok.
Are you sure you want to continue connecting (yes/no/[fingerprint])? yes
ok: [minecraft]

TASK [Try to connect] *****************************************************************************
ok: [minecraft]

TASK [Make persistent volume folder] **************************************************************
changed: [minecraft]

TASK [Mount persistent volume] ********************************************************************
changed: [minecraft]

TASK [Install JRE] ********************************************************************************
changed: [minecraft]

TASK [include vanilla variables] ******************************************************************
ok: [minecraft]

TASK [Make vanilla folder] ************************************************************************
changed: [minecraft]

TASK [Download Vanilla 1.16.4 jar] ****************************************************************
changed: [minecraft]

TASK [Agree to EULA] ******************************************************************************
changed: [minecraft]

TASK [Generate minecraft systemd service] *********************************************************
changed: [minecraft]

TASK [Start minecraft service] ********************************************************************
changed: [minecraft]

PLAY RECAP ****************************************************************************************
minecraft                  : ok=11   changed=8    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
```

```
nc -zv 159.89.95.238 25565
```

```
159.89.95.238 25565 open
```
