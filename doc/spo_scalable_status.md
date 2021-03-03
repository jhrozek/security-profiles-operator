# SPO scalable status

# Problem description
At the moment, the status of SPO profiles like seccomp or SELinux is per-profile. This presents several challenges:
 - it is impossible to reflect per-node status such as a profile failed to install on a single node or a node
   does not support the selected security profile
 - simply adding per-node attributes like a map might not scale, given very high number of nodes, the object might
   become too big for etcd's 1MB limit (see also https://github.com/kubernetes/kubernetes/issues/73324)
 - because the per-profile status is written to by several sources (each pod in a DaemonSet), the status might 
   appear as "flapping" as different pods reach different states at their own pace.

# Proposed solution overview
To address all the problems above, we should add a new status-like object that would represent profile
node status. This object (node status) would be owned by the profile object so that the node status is
garbage collected when the profile object is.

Design goals:
 - it must be possible to easily list all node statuses of one profile status
 - the profile status should display the overall profile status. This would be handy e.g. for column headers

# Implementation details
The current status of the seccomp profile looks like this:
```go
type SeccompProfileStatus struct {
	Path            string   `json:"path,omitempty"`
	Status          string   `json:"status,omitempty"`
	ActiveWorkloads []string `json:"activeWorkloads,omitempty"`
	// The path that should be provided to the `securityContext.seccompProfile.localhostProfile`
	// field of a Pod or container spec
	LocalhostProfile string `json:"localhostProfile,omitempty"`
}
```

The part if the profile status and the new `SeccompProfileNodeStatus` structure might look as follows:
```go
type SeccompProfileStatus struct {
    Path            string   `json:"path,omitempty"`
    Status          string   `json:"status,omitempty"`
    ActiveWorkloads []string `json:"activeWorkloads,omitempty"`
    // The path that should be provided to the `securityContext.seccompProfile.localhostProfile`
    // field of a Pod or container spec
    LocalhostProfile string `json:"localhostProfile,omitempty"`
}

type SeccompProfileNodeStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	
	NodeName        string   `json:"nodeName"`
	Status          string   `json:"status,omitempty"`
}
```

Since the reconcile loop runs in a pod which is part of a DS, it would first check if a node status for this node
exists and if not, create it with some initial state (`Initializing`?). The node status name would be set to
`profile+nodeName` and the node status would be labeled with the profile name. For readability, the node name
is included in the node status as well.

On status update, the controller would write the status for the node. Additionally, the profile status includes
an aggregated status as well, which would display the lowest common status of all node statuses - for example, 
if all node statuses are `Active` except for one in `Initializing`, the overall status would still be `Initializing`,
only when all statuses transition to `Active` does the aggregated status change to `Active` as well.

The SELinux controller could be amended along the same lines.

# Open questions
 - Several status attributes are inherently node-specific (`Path`, or even `ActiveWorkloads`), but it's
   unlikely that the admin would care about their value per node as opposed to knowing the status of a policy
   per node. Is that assumption correct?
 - Is it realistic that the `LocalhostProfile` or `Path` would be different on different nodes?
 - Events: Some events issues by the controllers are more specific to the node (e.g. failed to write a policy
   file to disk), some are better linked to the overall policy object (the profile contains a syntax error).
   We can:
      - split the events and issue those that pertain to the node for the node status object
      - keep issuing them for the profile itself, but try to add the node name as a key-value pair
        to the event messages. This is my preference as the events might be related to each other and
        it's easier to look at one place
