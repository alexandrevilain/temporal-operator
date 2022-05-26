<h1>API reference</h1>
<p>Package v1alpha1 contains API Schema definitions for the apps v1alpha1 API group</p>
Resource Types:
<ul class="simple"><li>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalCluster">TemporalCluster</a>
</li></ul>
<h3 id="apps.alexandrevilain.dev/v1alpha1.TemporalCluster">TemporalCluster
</h3>
<p>TemporalCluster is the Schema for the temporalclusters API</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br>
string</td>
<td>
<code>apps.alexandrevilain.dev/v1alpha1</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br>
string
</td>
<td>
<code>TemporalCluster</code>
</td>
</tr>
<tr>
<td>
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalClusterSpec">
TemporalClusterSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image defines the temporal server image the instance should use.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br>
<em>
string
</em>
</td>
<td>
<p>Version defines the temporal version the instance should run.</p>
</td>
</tr>
<tr>
<td>
<code>numHistoryShards</code><br>
<em>
int32
</em>
</td>
<td>
<p>NumHistoryShards is the desired number of history shards.
This field is immutable.</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalServicesSpec">
TemporalServicesSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>persistence</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalPersistenceSpec">
TemporalPersistenceSpec
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>datastores</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalDatastoreSpec">
[]TemporalDatastoreSpec
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalClusterStatus">
TemporalClusterStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="apps.alexandrevilain.dev/v1alpha1.DatastoreTLSSpec">DatastoreTLSSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalDatastoreSpec">TemporalDatastoreSpec</a>)
</p>
<p>DatastoreTLSSpec contains datastore TLS connections specifications.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>bool</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>certFileRef</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.SecretKeyReference">
SecretKeyReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>keyFileRef</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.SecretKeyReference">
SecretKeyReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>caFileRef</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.SecretKeyReference">
SecretKeyReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>enableHostVerification</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serverName</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="apps.alexandrevilain.dev/v1alpha1.DatastoreType">DatastoreType
(<code>string</code> alias)</h3>
<h3 id="apps.alexandrevilain.dev/v1alpha1.SQLSpec">SQLSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalDatastoreSpec">TemporalDatastoreSpec</a>)
</p>
<p>SQLSpec contains SQL datastore connections specifications.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>user</code><br>
<em>
string
</em>
</td>
<td>
<p>User is the username to be used for the connection.</p>
</td>
</tr>
<tr>
<td>
<code>pluginName</code><br>
<em>
string
</em>
</td>
<td>
<p>PluginName is the name of SQL plugin.</p>
</td>
</tr>
<tr>
<td>
<code>databaseName</code><br>
<em>
string
</em>
</td>
<td>
<p>DatabaseName is the name of SQL database to connect to.</p>
</td>
</tr>
<tr>
<td>
<code>connectAddr</code><br>
<em>
string
</em>
</td>
<td>
<p>ConnectAddr is the remote addr of the database.</p>
</td>
</tr>
<tr>
<td>
<code>connectProtocol</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ConnectProtocol is the protocol that goes with the ConnectAddr.</p>
</td>
</tr>
<tr>
<td>
<code>connectAttributes</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ConnectAttributes is a set of key-value attributes to be sent as part of connect data_source_name url</p>
</td>
</tr>
<tr>
<td>
<code>maxConns</code><br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>MaxConns the max number of connections to this datastore.</p>
</td>
</tr>
<tr>
<td>
<code>maxIdleConns</code><br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>MaxIdleConns is the max number of idle connections to this datastore.</p>
</td>
</tr>
<tr>
<td>
<code>maxConnLifetime</code><br>
<em>
time.Duration
</em>
</td>
<td>
<em>(Optional)</em>
<p>MaxConnLifetime is the maximum time a connection can be alive</p>
</td>
</tr>
<tr>
<td>
<code>taskScanPartitions</code><br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>TaskScanPartitions is the number of partitions to sequentially scan during ListTaskQueue operations.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="apps.alexandrevilain.dev/v1alpha1.SecretKeyReference">SecretKeyReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.DatastoreTLSSpec">DatastoreTLSSpec</a>, 
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalDatastoreSpec">TemporalDatastoreSpec</a>)
</p>
<p>SecretKeyReference contains enough information to locate the referenced Kubernetes Secret object in the same
namespace.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br>
<em>
string
</em>
</td>
<td>
<p>Name of the Secret.</p>
</td>
</tr>
<tr>
<td>
<code>key</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Key in the Secret.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="apps.alexandrevilain.dev/v1alpha1.ServiceSpec">ServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalServicesSpec">TemporalServicesSpec</a>)
</p>
<p>ServiceSpec contains a temporal service specifications.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>port</code><br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>membershipPort</code><br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="apps.alexandrevilain.dev/v1alpha1.TemporalClusterSpec">TemporalClusterSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalCluster">TemporalCluster</a>)
</p>
<p>TemporalClusterSpec defines the desired state of TemporalCluster.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image defines the temporal server image the instance should use.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br>
<em>
string
</em>
</td>
<td>
<p>Version defines the temporal version the instance should run.</p>
</td>
</tr>
<tr>
<td>
<code>numHistoryShards</code><br>
<em>
int32
</em>
</td>
<td>
<p>NumHistoryShards is the desired number of history shards.
This field is immutable.</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalServicesSpec">
TemporalServicesSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>persistence</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalPersistenceSpec">
TemporalPersistenceSpec
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>datastores</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalDatastoreSpec">
[]TemporalDatastoreSpec
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="apps.alexandrevilain.dev/v1alpha1.TemporalClusterStatus">TemporalClusterStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalCluster">TemporalCluster</a>)
</p>
<p>TemporalClusterStatus defines the observed state of TemporalCluster.</p>
<h3 id="apps.alexandrevilain.dev/v1alpha1.TemporalDatastoreSpec">TemporalDatastoreSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>TemporalDatastoreSpec contains temporal datastore specifications.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the datatstore.
It should be unique and will be referenced within the persitence spec.</p>
</td>
</tr>
<tr>
<td>
<code>sql</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.SQLSpec">
SQLSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SQL holds all connection parameters for SQL datastores.</p>
</td>
</tr>
<tr>
<td>
<code>passwordSecretRef</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.SecretKeyReference">
SecretKeyReference
</a>
</em>
</td>
<td>
<p>PasswordSecret is the reference to the secret holding the password.</p>
</td>
</tr>
<tr>
<td>
<code>tls</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.DatastoreTLSSpec">
DatastoreTLSSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS is an optional option to connect to the datastore using TLS.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="apps.alexandrevilain.dev/v1alpha1.TemporalPersistenceSpec">TemporalPersistenceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>TemporalPersistenceSpec contains temporal persistence specifications.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>defaultStore</code><br>
<em>
string
</em>
</td>
<td>
<p>DefaultStore is the name of the default data store to use.</p>
</td>
</tr>
<tr>
<td>
<code>visibilityStore</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>VisibilityStore is the name of the datastore to be used for visibility records.
If not set it defaults to the default store.</p>
</td>
</tr>
<tr>
<td>
<code>advancedVisibilityStore</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdvancedVisibilityStore is the name of the datastore to be used for visibility records</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="apps.alexandrevilain.dev/v1alpha1.TemporalServicesSpec">TemporalServicesSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#apps.alexandrevilain.dev/v1alpha1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>TemporalServicesSpec contains all temporal services specifications.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>frontend</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>history</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>matching</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>worker</code><br>
<em>
<a href="#apps.alexandrevilain.dev/v1alpha1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
