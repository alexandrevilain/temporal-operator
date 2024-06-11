<h1>API reference</h1>
<p>Package v1beta1 contains API Schema definitions for the v1beta1 API group</p>
Resource Types:
<ul class="simple"><li>
<a href="#temporal.io/v1beta1.TemporalCluster">TemporalCluster</a>
</li></ul>
<h3 id="temporal.io/v1beta1.TemporalCluster">TemporalCluster
</h3>
<p>TemporalCluster defines a temporal cluster deployment.</p>
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
<code>temporal.io/v1beta1</code>
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
<a href="#temporal.io/v1beta1.TemporalClusterSpec">
TemporalClusterSpec
</a>
</em>
</td>
<td>
<p>Specification of the desired behavior of the Temporal cluster.</p>
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
<p>Image defines the temporal server docker image the cluster should use for each services.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br>
<em>
github.com/alexandrevilain/temporal-operator/pkg/version.Version
</em>
</td>
<td>
<em>(Optional)</em>
<p>Version defines the temporal version the cluster to be deployed.
This version impacts the underlying persistence schemas versions.</p>
</td>
</tr>
<tr>
<td>
<code>log</code><br>
<em>
<a href="#temporal.io/v1beta1.LogSpec">
LogSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Log defines temporal cluster&rsquo;s logger configuration.</p>
</td>
</tr>
<tr>
<td>
<code>jobTtlSecondsAfterFinished</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>JobTTLSecondsAfterFinished is amount of time to keep job pods after jobs are completed.
Defaults to 300 seconds.</p>
</td>
</tr>
<tr>
<td>
<code>jobResources</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>JobResources allows set resources for setup/update jobs.</p>
</td>
</tr>
<tr>
<td>
<code>jobInitContainers</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>JobInitContainers adds a list of init containers to the setup&rsquo;s jobs.</p>
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
<a href="#temporal.io/v1beta1.ServicesSpec">
ServicesSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Services allows customizations for each temporal services deployment.</p>
</td>
</tr>
<tr>
<td>
<code>persistence</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalPersistenceSpec">
TemporalPersistenceSpec
</a>
</em>
</td>
<td>
<p>Persistence defines temporal persistence configuration.</p>
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
<p>An optional list of references to secrets in the same namespace
to use for pulling temporal images from registries.</p>
</td>
</tr>
<tr>
<td>
<code>ui</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalUISpec">
TemporalUISpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>UI allows configuration of the optional temporal web ui deployed alongside the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>admintools</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalAdminToolsSpec">
TemporalAdminToolsSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdminTools allows configuration of the optional admin tool pod deployed alongside the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>mTLS</code><br>
<em>
<a href="#temporal.io/v1beta1.MTLSSpec">
MTLSSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MTLS allows configuration of the network traffic encryption for the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>metrics</code><br>
<em>
<a href="#temporal.io/v1beta1.MetricsSpec">
MetricsSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metrics allows configuration of scraping endpoints for stats. prometheus or m3.</p>
</td>
</tr>
<tr>
<td>
<code>dynamicConfig</code><br>
<em>
<a href="#temporal.io/v1beta1.DynamicConfigSpec">
DynamicConfigSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DynamicConfig allows advanced configuration for the temporal cluster.</p>
</td>
</tr>
<tr>
<td>
<code>archival</code><br>
<em>
<a href="#temporal.io/v1beta1.ClusterArchivalSpec">
ClusterArchivalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Archival allows Workflow Execution Event Histories and Visibility data backups for the temporal cluster.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br>
<em>
<a href="#temporal.io/v1beta1.AuthorizationSpec">
AuthorizationSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization allows authorization configuration for the temporal cluster.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalClusterStatus">
TemporalClusterStatus
</a>
</em>
</td>
<td>
<p>Most recent observed status of the Temporal cluster.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ArchivalProvider">ArchivalProvider
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ClusterArchivalSpec">ClusterArchivalSpec</a>)
</p>
<p>ArchivalProvider contains the config for archivers.</p>
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
<code>filestore</code><br>
<em>
<a href="#temporal.io/v1beta1.FilestoreArchiver">
FilestoreArchiver
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>s3</code><br>
<em>
<a href="#temporal.io/v1beta1.S3Archiver">
S3Archiver
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>gcs</code><br>
<em>
<a href="#temporal.io/v1beta1.GCSArchiver">
GCSArchiver
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
<h3 id="temporal.io/v1beta1.ArchivalProviderKind">ArchivalProviderKind
(<code>string</code> alias)</h3>
<h3 id="temporal.io/v1beta1.ArchivalSpec">ArchivalSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ClusterArchivalSpec">ClusterArchivalSpec</a>, 
<a href="#temporal.io/v1beta1.TemporalNamespaceArchivalSpec">TemporalNamespaceArchivalSpec</a>)
</p>
<p>ArchivalSpec is the archival configuration for a particular persistence type (history or visibility).</p>
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled defines if the archival is enabled by default for all namespaces
or for a particular namespace (depends if it&rsquo;s for a TemporalCluster or a TemporalNamespace).</p>
</td>
</tr>
<tr>
<td>
<code>paused</code><br>
<em>
bool
</em>
</td>
<td>
<p>Paused defines if the archival is paused.</p>
</td>
</tr>
<tr>
<td>
<code>enableRead</code><br>
<em>
bool
</em>
</td>
<td>
<p>EnableRead allows temporal to read from the archived Event History.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br>
<em>
string
</em>
</td>
<td>
<p>Path is &hellip;</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.AuthorizationSpec">AuthorizationSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>AuthorizationSpec defines the specifications for authorization in the temporal cluster. It contains fields
that configure how JWT tokens are validated, how permissions are managed, and how claims are mapped.</p>
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
<code>jwtKeyProvider</code><br>
<em>
<a href="#temporal.io/v1beta1.AuthorizationSpecJWTKeyProvider">
AuthorizationSpecJWTKeyProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>JWTKeyProvider specifies the signing key provider used for validating JWT tokens.</p>
</td>
</tr>
<tr>
<td>
<code>permissionsClaimName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>PermissionsClaimName is the name of the claim within the JWT token that contains the user&rsquo;s permissions.</p>
</td>
</tr>
<tr>
<td>
<code>authorizer</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorizer defines the authorization mechanism to be used. It can be left as an empty string to
use a no-operation authorizer (noopAuthorizer), or set to &ldquo;default&rdquo; to use the temporal&rsquo;s default
authorizer (defaultAuthorizer).</p>
</td>
</tr>
<tr>
<td>
<code>claimMapper</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClaimMapper specifies the claim mapping mechanism used for handling JWT claims. Similar to the Authorizer,
it can be left as an empty string to use a no-operation claim mapper (noopClaimMapper), or set to &ldquo;default&rdquo;
to use the default JWT claim mapper (defaultJWTClaimMapper).</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.AuthorizationSpecJWTKeyProvider">AuthorizationSpecJWTKeyProvider
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.AuthorizationSpec">AuthorizationSpec</a>)
</p>
<p>AuthorizationSpecJWTKeyProvider defines the configuration for a JWT key provider within the AuthorizationSpec.
It specifies where to source the JWT keys from and how often they should be refreshed.</p>
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
<code>keySourceURIs</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>KeySourceURIs is a list of URIs where the JWT signing keys can be obtained. These URIs are used by the
authorization system to fetch the public keys necessary for validating JWT tokens.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RefreshInterval defines the time interval at which temporal should refresh the JWT signing keys from
the specified URIs.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.CassandraConsistencySpec">CassandraConsistencySpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.CassandraSpec">CassandraSpec</a>)
</p>
<p>CassandraConsistencySpec sets the consistency level for regular &amp; serial queries to Cassandra.</p>
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
<code>consistency</code><br>
<em>
<a href="https://pkg.go.dev/github.com/gocql/gocql#Consistency">
github.com/gocql/gocql.Consistency
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Consistency sets the default consistency level.
Values identical to gocql Consistency values. (defaults to LOCAL_QUORUM if not set).</p>
</td>
</tr>
<tr>
<td>
<code>serialConsistency</code><br>
<em>
<a href="https://pkg.go.dev/github.com/gocql/gocql#SerialConsistencyy">
github.com/gocql/gocql.SerialConsistency
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SerialConsistency sets the consistency for the serial prtion of queries. Values identical to gocql SerialConsistency values.
(defaults to LOCAL_SERIAL if not set)</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.CassandraSpec">CassandraSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DatastoreSpec">DatastoreSpec</a>)
</p>
<p>CassandraSpec contains cassandra datastore connections specifications.</p>
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
<code>hosts</code><br>
<em>
[]string
</em>
</td>
<td>
<p>Hosts is a list of cassandra endpoints.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br>
<em>
int
</em>
</td>
<td>
<p>Port is the cassandra port used for connection by gocql client.</p>
</td>
</tr>
<tr>
<td>
<code>user</code><br>
<em>
string
</em>
</td>
<td>
<p>User is the cassandra user used for authentication by gocql client.</p>
</td>
</tr>
<tr>
<td>
<code>keyspace</code><br>
<em>
string
</em>
</td>
<td>
<p>Keyspace is the cassandra keyspace.</p>
</td>
</tr>
<tr>
<td>
<code>datacenter</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Datacenter is the data center filter arg for cassandra.</p>
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
<p>MaxConns is the max number of connections to this datastore for a single keyspace.</p>
</td>
</tr>
<tr>
<td>
<code>connectTimeout</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ConnectTimeout is a timeout for initial dial to cassandra server.</p>
</td>
</tr>
<tr>
<td>
<code>consistency</code><br>
<em>
<a href="#temporal.io/v1beta1.CassandraConsistencySpec">
CassandraConsistencySpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Consistency configuration.</p>
</td>
</tr>
<tr>
<td>
<code>disableInitialHostLookup</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DisableInitialHostLookup instructs the gocql client to connect only using the supplied hosts.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.CertificatesDurationSpec">CertificatesDurationSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.MTLSSpec">MTLSSpec</a>)
</p>
<p>CertificatesDurationSpec defines parameters for the temporal mTLS certificates duration.</p>
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
<code>rootCACertificate</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RootCACertificate is the &lsquo;duration&rsquo; (i.e. lifetime) of the Root CA Certificate.
It defaults to 10 years.</p>
</td>
</tr>
<tr>
<td>
<code>intermediateCAsCertificates</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IntermediateCACertificates is the &lsquo;duration&rsquo; (i.e. lifetime) of the intermediate CAs Certificates.
It defaults to 5 years.</p>
</td>
</tr>
<tr>
<td>
<code>clientCertificates</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClientCertificates is the &lsquo;duration&rsquo; (i.e. lifetime) of the client certificates.
It defaults to 1 year.</p>
</td>
</tr>
<tr>
<td>
<code>frontendCertificate</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>FrontendCertificate is the &lsquo;duration&rsquo; (i.e. lifetime) of the frontend certificate.
It defaults to 1 year.</p>
</td>
</tr>
<tr>
<td>
<code>internodeCertificate</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>InternodeCertificate is the &lsquo;duration&rsquo; (i.e. lifetime) of the internode certificate.
It defaults to 1 year.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ClusterArchivalSpec">ClusterArchivalSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>ClusterArchivalSpec is the configuration for cluster-wide archival config.</p>
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled defines if the archival is enabled for the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>provider</code><br>
<em>
<a href="#temporal.io/v1beta1.ArchivalProvider">
ArchivalProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Provider defines the archival provider for the cluster.
The same provider is used for both history and visibility,
but some config can be changed using spec.archival.[history|visibility].config.</p>
</td>
</tr>
<tr>
<td>
<code>history</code><br>
<em>
<a href="#temporal.io/v1beta1.ArchivalSpec">
ArchivalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>History is the default config for the history archival.</p>
</td>
</tr>
<tr>
<td>
<code>visibility</code><br>
<em>
<a href="#temporal.io/v1beta1.ArchivalSpec">
ArchivalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Visibility is the default config for visibility archival.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ConstrainedValue">ConstrainedValue
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DynamicConfigSpec">DynamicConfigSpec</a>)
</p>
<p>ConstrainedValue is an alias for temporal&rsquo;s dynamicconfig.ConstrainedValue.</p>
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
<code>constraints</code><br>
<em>
<a href="#temporal.io/v1beta1.Constraints">
Constraints
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Constraints describe under what conditions a ConstrainedValue should be used.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<p>Value is the value for the configuration key.
The type of the Value field depends on the key.
Acceptable types will be one of: int, float64, bool, string, map[string]any, time.Duration</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.Constraints">Constraints
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ConstrainedValue">ConstrainedValue</a>)
</p>
<p>Constraints is an alias for temporal&rsquo;s dynamicconfig.Constraints.
It describes under what conditions a ConstrainedValue should be used.</p>
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
<code>namespace</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>namespaceId</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>taskQueueName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>taskQueueType</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>shardId</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>taskType</code><br>
<em>
string
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
<h3 id="temporal.io/v1beta1.DatastoreSpec">DatastoreSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalPersistenceSpec">TemporalPersistenceSpec</a>)
</p>
<p>DatastoreSpec contains temporal datastore specifications.</p>
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
<em>(Optional)</em>
<p>Name is the name of the datastore.
It should be unique and will be referenced within the persistence spec.
Defaults to &ldquo;default&rdquo; for default sore, &ldquo;visibility&rdquo; for visibility store,
&ldquo;secondaryVisibility&rdquo; for secondary visibility store and
&ldquo;advancedVisibility&rdquo; for advanced visibility store.</p>
</td>
</tr>
<tr>
<td>
<code>sql</code><br>
<em>
<a href="#temporal.io/v1beta1.SQLSpec">
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
<code>elasticsearch</code><br>
<em>
<a href="#temporal.io/v1beta1.ElasticsearchSpec">
ElasticsearchSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Elasticsearch holds all connection parameters for Elasticsearch datastores.</p>
</td>
</tr>
<tr>
<td>
<code>cassandra</code><br>
<em>
<a href="#temporal.io/v1beta1.CassandraSpec">
CassandraSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Cassandra holds all connection parameters for Cassandra datastore.
Note that cassandra is now deprecated for visibility store.</p>
</td>
</tr>
<tr>
<td>
<code>passwordSecretRef</code><br>
<em>
<a href="#temporal.io/v1beta1.SecretKeyReference">
SecretKeyReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PasswordSecret is the reference to the secret holding the password.</p>
</td>
</tr>
<tr>
<td>
<code>tls</code><br>
<em>
<a href="#temporal.io/v1beta1.DatastoreTLSSpec">
DatastoreTLSSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS is an optional option to connect to the datastore using TLS.</p>
</td>
</tr>
<tr>
<td>
<code>skipCreate</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>SkipCreate instructs the operator to skip creating the database for SQL datastores or to skip creating keyspace for Cassandra. Use this option if your database or keyspace has already been provisioned by an administrator.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.DatastoreStatus">DatastoreStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalPersistenceStatus">TemporalPersistenceStatus</a>)
</p>
<p>DatastoreStatus contains the current status of a datastore.</p>
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
<code>created</code><br>
<em>
bool
</em>
</td>
<td>
<p>Created indicates if the database or keyspace has been created.</p>
</td>
</tr>
<tr>
<td>
<code>setup</code><br>
<em>
bool
</em>
</td>
<td>
<p>Setup indicates if tables have been set up.</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br>
<em>
<a href="#temporal.io/v1beta1.DatastoreType">
DatastoreType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Type indicates the datastore type.</p>
</td>
</tr>
<tr>
<td>
<code>schemaVersion</code><br>
<em>
github.com/alexandrevilain/temporal-operator/pkg/version.Version
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchemaVersion report the current schema version.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.DatastoreTLSSpec">DatastoreTLSSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DatastoreSpec">DatastoreSpec</a>)
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<p>Enabled defines if the cluster should use a TLS connection to connect to the datastore.</p>
</td>
</tr>
<tr>
<td>
<code>certFileRef</code><br>
<em>
<a href="#temporal.io/v1beta1.SecretKeyReference">
SecretKeyReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CertFileRef is a reference to a secret containing the cert file.</p>
</td>
</tr>
<tr>
<td>
<code>keyFileRef</code><br>
<em>
<a href="#temporal.io/v1beta1.SecretKeyReference">
SecretKeyReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>KeyFileRef is a reference to a secret containing the key file.</p>
</td>
</tr>
<tr>
<td>
<code>caFileRef</code><br>
<em>
<a href="#temporal.io/v1beta1.SecretKeyReference">
SecretKeyReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CaFileRef is a reference to a secret containing the ca file.</p>
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
<p>EnableHostVerification defines if the hostname should be verified when connecting to the datastore.</p>
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
<em>(Optional)</em>
<p>ServerName the datastore should present.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.DatastoreType">DatastoreType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DatastoreStatus">DatastoreStatus</a>)
</p>
<h3 id="temporal.io/v1beta1.DeploymentOverride">DeploymentOverride
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ServiceSpecOverride">ServiceSpecOverride</a>)
</p>
<p>DeploymentOverride provides the ability to override a Deployment.</p>
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
<code>metadata</code><br>
<em>
<a href="#temporal.io/v1beta1.ObjectMetaOverride">
ObjectMetaOverride
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#temporal.io/v1beta1.DeploymentOverrideSpec">
DeploymentOverrideSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specification of the desired behavior of the Deployment.</p>
<br/>
<br/>
<table>
</table>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.DeploymentOverrideSpec">DeploymentOverrideSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DeploymentOverride">DeploymentOverride</a>)
</p>
<p>DeploymentOverrideSpec provides the ability to override a Deployment Spec.
It&rsquo;s a subset of fields included in k8s.io/api/apps/v1.DeploymentSpec.</p>
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
<code>template</code><br>
<em>
<a href="#temporal.io/v1beta1.PodTemplateSpecOverride">
PodTemplateSpecOverride
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Template describes the pods that will be created.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.DynamicConfigSpec">DynamicConfigSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>DynamicConfigSpec is the configuration for temporal dynamic config.</p>
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
<code>pollInterval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PollInterval defines how often the config should be updated by checking provided values.
Defaults to 10s.</p>
</td>
</tr>
<tr>
<td>
<code>values</code><br>
<em>
<a href="#temporal.io/v1beta1.ConstrainedValue">
map[string][]./api/v1beta1.ConstrainedValue
</a>
</em>
</td>
<td>
<p>Values contains all dynamic config keys and their constrained values.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ElasticsearchIndices">ElasticsearchIndices
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ElasticsearchSpec">ElasticsearchSpec</a>)
</p>
<p>ElasticsearchIndices holds index names.</p>
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
<code>visibility</code><br>
<em>
string
</em>
</td>
<td>
<p>Visibility defines visibility&rsquo;s index name.</p>
</td>
</tr>
<tr>
<td>
<code>secondaryVisibility</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecondaryVisibility defines secondary visibility&rsquo;s index name.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ElasticsearchSpec">ElasticsearchSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DatastoreSpec">DatastoreSpec</a>)
</p>
<p>ElasticsearchSpec contains Elasticsearch datastore connections specifications.</p>
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
<code>version</code><br>
<em>
string
</em>
</td>
<td>
<p>Version defines the elasticsearch version.</p>
</td>
</tr>
<tr>
<td>
<code>url</code><br>
<em>
string
</em>
</td>
<td>
<p>URL is the connection url to connect to the instance.</p>
</td>
</tr>
<tr>
<td>
<code>username</code><br>
<em>
string
</em>
</td>
<td>
<p>Username is the username to be used for the connection.</p>
</td>
</tr>
<tr>
<td>
<code>indices</code><br>
<em>
<a href="#temporal.io/v1beta1.ElasticsearchIndices">
ElasticsearchIndices
</a>
</em>
</td>
<td>
<p>Indices holds visibility index names.</p>
</td>
</tr>
<tr>
<td>
<code>logLevel</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>LogLevel defines the temporal cluster&rsquo;s es client logger level.</p>
</td>
</tr>
<tr>
<td>
<code>closeIdleConnectionsInterval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CloseIdleConnectionsInterval is the max duration a connection stay open while idle.</p>
</td>
</tr>
<tr>
<td>
<code>enableSniff</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>EnableSniff enables or disables sniffer on the temporal cluster&rsquo;s es client.</p>
</td>
</tr>
<tr>
<td>
<code>enableHealthcheck</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>EnableHealthcheck enables or disables healthcheck on the temporal cluster&rsquo;s es client.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.FilestoreArchiver">FilestoreArchiver
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ArchivalProvider">ArchivalProvider</a>)
</p>
<p>FilestoreArchiver is the file store archival provider configuration.</p>
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
<code>filePermissions</code><br>
<em>
string
</em>
</td>
<td>
<p>FilePermissions sets the file permissions of the archived files.
It&rsquo;s recommend to leave it empty and use the default value of &ldquo;0666&rdquo; to avoid read/write issues.</p>
</td>
</tr>
<tr>
<td>
<code>dirPermissions</code><br>
<em>
string
</em>
</td>
<td>
<p>DirPermissions sets the directory permissions of the archive directory.
It&rsquo;s recommend to leave it empty and use the default value of &ldquo;0766&rdquo; to avoid read/write issues.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.FrontendMTLSSpec">FrontendMTLSSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.MTLSSpec">MTLSSpec</a>)
</p>
<p>FrontendMTLSSpec defines parameters for the temporal encryption in transit with mTLS.</p>
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled defines if the operator should enable mTLS for cluster&rsquo;s public endpoints.</p>
</td>
</tr>
<tr>
<td>
<code>extraDnsNames</code><br>
<em>
[]string
</em>
</td>
<td>
<p>ExtraDNSNames is a list of additional DNS names associated with the TemporalCluster.
These DNS names can be used for accessing the TemporalCluster from external services.
The DNS names specified here will be added to the TLS certificate for secure communication.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.GCSArchiver">GCSArchiver
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ArchivalProvider">ArchivalProvider</a>)
</p>
<p>GCSArchiver is the GCS archival provider configuration.</p>
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
<code>credentialsRef</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>SecretAccessKeyRef is the secret key selector containing Google Cloud Storage credentials file.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.InternalFrontendServiceSpec">InternalFrontendServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ServicesSpec">ServicesSpec</a>)
</p>
<p>InternalFrontendServiceSpec contains temporal internal frontend service specifications.</p>
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
<code>ServiceSpec</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>ServiceSpec</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled defines if we want to spawn the internal frontend service.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.InternodeMTLSSpec">InternodeMTLSSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.MTLSSpec">MTLSSpec</a>)
</p>
<p>InternodeMTLSSpec defines parameters for the temporal encryption in transit with mTLS.</p>
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled defines if the operator should enable mTLS for network between cluster nodes.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.LogSpec">LogSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>LogSpec contains the temporal logging configuration.</p>
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
<code>stdout</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Stdout is true if the output needs to goto standard out; default is stderr.</p>
</td>
</tr>
<tr>
<td>
<code>level</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Level is the desired log level; see colocated zap_logger.go::parseZapLevel()</p>
</td>
</tr>
<tr>
<td>
<code>outputFile</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>OutputFile is the path to the log output file.</p>
</td>
</tr>
<tr>
<td>
<code>format</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Format determines the format of each log file printed to the output.
Use &ldquo;console&rdquo; if you want stack traces to appear on multiple lines.</p>
</td>
</tr>
<tr>
<td>
<code>development</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Development determines whether the logger is run in Development (== Test) or in
Production mode.  Default is Production.  Production-stage disables panics from
DPanic logging.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.MTLSProvider">MTLSProvider
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.MTLSSpec">MTLSSpec</a>)
</p>
<p>MTLSProvider is the enum for support mTLS provider.</p>
<h3 id="temporal.io/v1beta1.MTLSSpec">MTLSSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>MTLSSpec defines parameters for the temporal encryption in transit with mTLS.</p>
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
<code>provider</code><br>
<em>
<a href="#temporal.io/v1beta1.MTLSProvider">
MTLSProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Provider defines the tool used to manage mTLS certificates.</p>
</td>
</tr>
<tr>
<td>
<code>internode</code><br>
<em>
<a href="#temporal.io/v1beta1.InternodeMTLSSpec">
InternodeMTLSSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Internode allows configuration of the internode traffic encryption.
Useless if mTLS provider is not cert-manager.</p>
</td>
</tr>
<tr>
<td>
<code>frontend</code><br>
<em>
<a href="#temporal.io/v1beta1.FrontendMTLSSpec">
FrontendMTLSSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Frontend allows configuration of the frontend&rsquo;s public endpoint traffic encryption.
Useless if mTLS provider is not cert-manager.</p>
</td>
</tr>
<tr>
<td>
<code>certificatesDuration</code><br>
<em>
<a href="#temporal.io/v1beta1.CertificatesDurationSpec">
CertificatesDurationSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CertificatesDuration allows configuration of maximum certificates lifetime.
Useless if mTLS provider is not cert-manager.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RefreshInterval defines interval between refreshes of certificates in the cluster components.
Defaults to 1 hour.
Useless if mTLS provider is not cert-manager.</p>
</td>
</tr>
<tr>
<td>
<code>renewBefore</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RenewBefore is defines how long before the currently issued certificate&rsquo;s expiry
cert-manager should renew the certificate. The default is <sup>2</sup>&frasl;<sub>3</sub> of the
issued certificate&rsquo;s duration. Minimum accepted value is 5 minutes.
Useless if mTLS provider is not cert-manager.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.MetricsSpec">MetricsSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>MetricsSpec determines parameters for configuring metrics endpoints.</p>
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<p>Enabled defines if the operator should enable metrics exposition on temporal components.</p>
</td>
</tr>
<tr>
<td>
<code>excludeTags</code><br>
<em>
map[string][]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExcludeTags is a map from tag name string to tag values string list.
Each value present in keys will have relevant tag value replaced with &ldquo;_tag<em>excluded</em>&rdquo;
Each value in values list will white-list tag values to be reported as usual.</p>
</td>
</tr>
<tr>
<td>
<code>perUnitHistogramBoundaries</code><br>
<em>
map[string][]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>PerUnitHistogramBoundaries defines the default histogram bucket boundaries.
Configuration of histogram boundaries for given metric unit.</p>
<p>Supported values:
- &ldquo;dimensionless&rdquo;
- &ldquo;milliseconds&rdquo;
- &ldquo;bytes&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Prefix sets the prefix to all outgoing metrics</p>
</td>
</tr>
<tr>
<td>
<code>prometheus</code><br>
<em>
<a href="#temporal.io/v1beta1.PrometheusSpec">
PrometheusSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Prometheus reporter configuration.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ObjectMetaOverride">ObjectMetaOverride
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DeploymentOverride">DeploymentOverride</a>, 
<a href="#temporal.io/v1beta1.PodTemplateSpecOverride">PodTemplateSpecOverride</a>, 
<a href="#temporal.io/v1beta1.TemporalUISpec">TemporalUISpec</a>)
</p>
<p>ObjectMetaOverride provides the ability to override an object metadata.
It&rsquo;s a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta.</p>
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
<code>labels</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Map of string keys and values that can be used to organize and categorize
(scope and select) objects.</p>
</td>
</tr>
<tr>
<td>
<code>annotations</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations is an unstructured key value map stored with a resource that may be
set by external tools to store and retrieve arbitrary metadata.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.PodTemplateSpecOverride">PodTemplateSpecOverride
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DeploymentOverrideSpec">DeploymentOverrideSpec</a>)
</p>
<p>PodTemplateSpecOverride provides the ability to override a pod template spec.
It&rsquo;s a subset of the fields included in k8s.io/api/core/v1.PodTemplateSpec.</p>
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
<code>metadata</code><br>
<em>
<a href="#temporal.io/v1beta1.ObjectMetaOverride">
ObjectMetaOverride
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specification of the desired behavior of the pod.</p>
<br/>
<br/>
<table>
</table>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.PrometheusScrapeConfig">PrometheusScrapeConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.PrometheusSpec">PrometheusSpec</a>)
</p>
<p>PrometheusScrapeConfig is the configuration for making prometheus scrape components metrics.</p>
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
<code>annotations</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations defines if the operator should add prometheus scrape annotations to the services pods.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitor</code><br>
<em>
<a href="#temporal.io/v1beta1.PrometheusScrapeConfigServiceMonitor">
PrometheusScrapeConfigServiceMonitor
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
<h3 id="temporal.io/v1beta1.PrometheusScrapeConfigServiceMonitor">PrometheusScrapeConfigServiceMonitor
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.PrometheusScrapeConfig">PrometheusScrapeConfig</a>)
</p>
<p>PrometheusScrapeConfigServiceMonitor is the configuration for prometheus operator ServiceMonitor.</p>
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled defines if the operator should create a ServiceMonitor for each services.</p>
</td>
</tr>
<tr>
<td>
<code>labels</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels adds extra labels to the ServiceMonitor.</p>
</td>
</tr>
<tr>
<td>
<code>override</code><br>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.ServiceMonitorSpec">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.ServiceMonitorSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Override allows customization of the created ServiceMonitor.
All fields can be overwritten except &ldquo;endpoints&rdquo;, &ldquo;selector&rdquo; and &ldquo;namespaceSelector&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>metricRelabelings</code><br>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.RelabelConfig">
[]github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.RelabelConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MetricRelabelConfigs to apply to samples before ingestion.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.PrometheusSpec">PrometheusSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.MetricsSpec">MetricsSpec</a>)
</p>
<p>PrometheusSpec is the configuration for prometheus reporter.</p>
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
<code>listenAddress</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Deprecated. Address for prometheus to serve metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>listenPort</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>ListenPort for prometheus to serve metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfig</code><br>
<em>
<a href="#temporal.io/v1beta1.PrometheusScrapeConfig">
PrometheusScrapeConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScrapeConfig is the prometheus scrape configuration.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.RetryPolicy">RetryPolicy
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleWorkflowAction">ScheduleWorkflowAction</a>)
</p>
<p>RetryPolicy defines how retries ought to be handled, usable by both workflows and activities.</p>
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
<code>initialInterval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Interval of the first retry. If retryBackoffCoefficient is 1.0 then it is used for all retries.</p>
</td>
</tr>
<tr>
<td>
<code>backoffCoefficient</code><br>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<em>(Optional)</em>
<p>Coefficient used to calculate the next retry interval.
The next retry interval is previous interval multiplied by the coefficient.
Must be 1 or larger.</p>
</td>
</tr>
<tr>
<td>
<code>maximumInterval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Maximum interval between retries. Exponential backoff leads to interval increase.
This value is the cap of the increase. Default is 100x of the initial interval.</p>
</td>
</tr>
<tr>
<td>
<code>maximumAttempts</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Maximum number of attempts. When exceeded the retries stop even if not expired yet.
1 disables retries. 0 means unlimited (up to the timeouts).</p>
</td>
</tr>
<tr>
<td>
<code>nonRetryableErrorTypes</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Non-Retryable errors types. Will stop retrying if the error type matches this list. Note that
this is not a substring match, the error <em>type</em> (not message) must match exactly.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.S3Archiver">S3Archiver
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ArchivalProvider">ArchivalProvider</a>)
</p>
<p>S3Archiver is the S3 archival provider configuration.</p>
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
<code>region</code><br>
<em>
string
</em>
</td>
<td>
<p>Region is the aws s3 region.</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use Endpoint if you want to use s3-compatible object storage.</p>
</td>
</tr>
<tr>
<td>
<code>roleName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use RoleName if you want the temporal service account
to assume an AWS Identity and Access Management (IAM) role.</p>
</td>
</tr>
<tr>
<td>
<code>credentials</code><br>
<em>
<a href="#temporal.io/v1beta1.S3Credentials">
S3Credentials
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use credentials if you want to use aws credentials from secret.</p>
</td>
</tr>
<tr>
<td>
<code>s3ForcePathStyle</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use s3ForcePathStyle if you want to use s3 path style.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.S3Credentials">S3Credentials
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.S3Archiver">S3Archiver</a>)
</p>
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
<code>accessKeyIdRef</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AccessKeyIDRef is the secret key selector containing AWS access key ID.</p>
</td>
</tr>
<tr>
<td>
<code>secretKeyRef</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>SecretAccessKeyRef is the secret key selector containing AWS secret access key.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.SQLSpec">SQLSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DatastoreSpec">DatastoreSpec</a>)
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
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
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
<tr>
<td>
<code>gcpServiceAccount</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>GCPServiceAccount is the service account to use to authenticate with GCP CloudSQL.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.Schedule">Schedule
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalScheduleSpec">TemporalScheduleSpec</a>)
</p>
<p>Schedule contains all fields related to a schedule.</p>
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
<code>action</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleAction">
ScheduleAction
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleSpec">
ScheduleSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>calendars</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">
[]ScheduleCalendarSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Calendars represents calendar-based specifications of times.</p>
</td>
</tr>
<tr>
<td>
<code>intervals</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleIntervalSpec">
[]ScheduleIntervalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Intervals represents interval-based specifications of times.</p>
</td>
</tr>
<tr>
<td>
<code>crons</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Crons are cron based specifications of times.
Crons is provided for easy migration from legacy Cron Workflows. For new
use cases, we recommend using ScheduleSpec.Calendars or ScheduleSpec.
Intervals for readability and maintainability. Once a schedule is created all
expressions in Crons will be translated to ScheduleSpec.Calendars on the server.</p>
<p>For example, <code>0 12 * * MON-WED,FRI</code> is every M/Tu/W/F at noon</p>
<p>The string can have 5, 6, or 7 fields, separated by spaces, and they are interpreted in the
same way as a ScheduleCalendarSpec:</p>
<pre><code>- 5 fields:         Minute, Hour, DayOfMonth, Month, DayOfWeek
- 6 fields:         Minute, Hour, DayOfMonth, Month, DayOfWeek, Year
- 7 fields: Second, Minute, Hour, DayOfMonth, Month, DayOfWeek, Year
</code></pre>
<p>Notes:
- If Year is not given, it defaults to *.
- If Second is not given, it defaults to 0.
- Shorthands @yearly, @monthly, @weekly, @daily, and @hourly are also
accepted instead of the 5-7 time fields.
- @every <interval>[/<phase>] is accepted and gets compiled into an
IntervalSpec instead. <interval> and <phase> should be a decimal integer
with a unit suffix s, m, h, or d.
- Optionally, the string can be preceded by CRON_TZ=<time zone name> or
TZ=<time zone name>, which will get copied to ScheduleSpec.TimeZoneName. (In which case the ScheduleSpec.TimeZone field should be left empty.)
- Optionally, &ldquo;#&rdquo; followed by a comment can appear at the end of the string.
- Note that the special case that some cron implementations have for
treating DayOfMonth and DayOfWeek as &ldquo;or&rdquo; instead of &ldquo;and&rdquo; when both
are set is not implemented.</p>
</td>
</tr>
<tr>
<td>
<code>excludeCalendars</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">
[]ScheduleCalendarSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExcludeCalendars defines any matching times that will be skipped.</p>
<p>All fields of the ScheduleCalendarSpec including seconds must match a time for the time to be skipped.</p>
</td>
</tr>
<tr>
<td>
<code>startAt</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartAt represents the start of the schedule. Any times before <code>startAt</code> will be skipped.
Together, <code>startAt</code> and <code>endAt</code> make an inclusive interval.
Defaults to the beginning of time.
For example: 2024-05-13T00:00:00Z</p>
</td>
</tr>
<tr>
<td>
<code>endAt</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EndAt represents the end of the schedule. Any times after <code>endAt</code> will be skipped.
Defaults to the end of time.
For example: 2024-05-13T00:00:00Z</p>
</td>
</tr>
<tr>
<td>
<code>jitter</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Jitter represents a duration that is used to apply a jitter to scheduled times.
All times will be incremented by a random value from 0 to this amount of jitter, capped
by the time until the next schedule.
Defaults to 0.</p>
</td>
</tr>
<tr>
<td>
<code>timezoneName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TimeZoneName represents the IANA time zone name, for example <code>US/Pacific</code>.</p>
<p>The definition will be loaded by Temporal Server from the environment it runs in.</p>
<p>Calendar spec matching is based on literal matching of the clock time
with no special handling of DST: if you write a calendar spec that fires
at 2:30am and specify a time zone that follows DST, that action will not
be triggered on the day that has no 2:30am. Similarly, an action that
fires at 1:30am will be triggered twice on the day that has two 1:30s.</p>
<p>Note: No actions are taken on leap-seconds (e.g. 23:59:60 UTC).
Defaults to UTC.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>policy</code><br>
<em>
<a href="#temporal.io/v1beta1.SchedulePolicies">
SchedulePolicies
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>state</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleState">
ScheduleState
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
<h3 id="temporal.io/v1beta1.ScheduleAction">ScheduleAction
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.Schedule">Schedule</a>)
</p>
<p>ScheduleAction contains the actions that the schedule should perform.</p>
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
<code>workflow</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleWorkflowAction">
ScheduleWorkflowAction
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
<h3 id="temporal.io/v1beta1.ScheduleCalendarSpec">ScheduleCalendarSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleSpec">ScheduleSpec</a>)
</p>
<p>ScheduleCalendarSpec is an event specification relative to the calendar, similar to a traditional cron specification.
A timestamp matches if at least one range of each field matches the
corresponding fields of the timestamp, except for year: if year is missing,
that means all years match. For all fields besides year, at least one Range must be present to match anything.</p>
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
<code>second</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleSecondMinuteRange">
[]ScheduleSecondMinuteRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Second range to match (0-59).
Defaults to 0.</p>
</td>
</tr>
<tr>
<td>
<code>minute</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleSecondMinuteRange">
[]ScheduleSecondMinuteRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Minute range to match (0-59).
Defaults to 0.</p>
</td>
</tr>
<tr>
<td>
<code>hour</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleHourRange">
[]ScheduleHourRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Hour range to match (0-23).
Defaults to 0.</p>
</td>
</tr>
<tr>
<td>
<code>dayOfMonth</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleDayOfMonthRange">
[]ScheduleDayOfMonthRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DayOfMonth range to match (1-31)
Defaults to match all days.</p>
</td>
</tr>
<tr>
<td>
<code>month</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleMonthRange">
[]ScheduleMonthRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Month range to match (1-12).
Defaults to match all months.</p>
</td>
</tr>
<tr>
<td>
<code>year</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleYearRange">
[]ScheduleYearRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Year range to match.
Defaults to match all years.</p>
</td>
</tr>
<tr>
<td>
<code>dayOfWeek</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleDayOfWeekRange">
[]ScheduleDayOfWeekRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DayOfWeek range to match (0-6; 0 is Sunday)
Defaults to match all days of the week.</p>
</td>
</tr>
<tr>
<td>
<code>comment</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Comment describes the intention of this schedule.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleDayOfMonthRange">ScheduleDayOfMonthRange
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">ScheduleCalendarSpec</a>)
</p>
<p>If end &lt; start, then end is interpreted as
equal to start. This means you can use a Range with start set to a value, and
end and step unset to represent a single value.</p>
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
<code>start</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Start of the range (inclusive).
Defaults to 1.</p>
</td>
</tr>
<tr>
<td>
<code>end</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>End of the range (inclusive).
Defaults to start.</p>
</td>
</tr>
<tr>
<td>
<code>step</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Step to be take between each value.
Defaults to 1.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleDayOfWeekRange">ScheduleDayOfWeekRange
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">ScheduleCalendarSpec</a>)
</p>
<p>If end &lt; start, then end is interpreted as
equal to start. This means you can use a Range with start set to a value, and
end and step unset to represent a single value.</p>
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
<code>start</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Start of the range (inclusive).
Defaults to 0.</p>
</td>
</tr>
<tr>
<td>
<code>end</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>End of the range (inclusive).
Defaults to start.</p>
</td>
</tr>
<tr>
<td>
<code>step</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Step to be take between each value.
Defaults to 1.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleHourRange">ScheduleHourRange
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">ScheduleCalendarSpec</a>)
</p>
<p>If end &lt; start, then end is interpreted as
equal to start. This means you can use a Range with start set to a value, and
end and step unset to represent a single value.</p>
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
<code>start</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Start of the range (inclusive).
Defaults to 0.</p>
</td>
</tr>
<tr>
<td>
<code>end</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>End of the range (inclusive).
Defaults to start.</p>
</td>
</tr>
<tr>
<td>
<code>step</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Step to be take between each value.
Defaults to 1.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleIntervalSpec">ScheduleIntervalSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleSpec">ScheduleSpec</a>)
</p>
<p>ScheduleIntervalSpec matches times that can be expressed as:</p>
<pre><code>Epoch + (n * every) + offset
where n is all integers ≥ 0.
</code></pre>
<p>For example, an <code>every</code> of 1 hour with <code>offset</code> of zero would match every hour, on the hour. The same <code>every</code> but an <code>offset</code>
of 19 minutes would match every <code>xx:19:00</code>. An <code>every</code> of 28 days with <code>offset</code> zero would match <code>2022-02-17T00:00:00Z</code>
(among other times). The same <code>every</code> with <code>offset</code> of 3 days, 5 hours, and 23 minutes would match <code>2022-02-20T05:23:00Z</code>
instead.</p>
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
<code>every</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<p>Every describes the period to repeat the interval.</p>
</td>
</tr>
<tr>
<td>
<code>offset</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Offset is a fixed offset added to the intervals period.
Defaults to 0.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleMonthRange">ScheduleMonthRange
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">ScheduleCalendarSpec</a>)
</p>
<p>If end &lt; start, then end is interpreted as
equal to start. This means you can use a Range with start set to a value, and
end and step unset to represent a single value.</p>
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
<code>start</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Start of the range (inclusive).
Defaults to 1.</p>
</td>
</tr>
<tr>
<td>
<code>end</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>End of the range (inclusive).
Defaults to start.</p>
</td>
</tr>
<tr>
<td>
<code>step</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Step to be take between each value.
Defaults to 1.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleOverlapPolicy">ScheduleOverlapPolicy
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.SchedulePolicies">SchedulePolicies</a>)
</p>
<p>Overlap controls what happens when an Action would be started by a
Schedule at the same time that an older Action is still running.</p>
<p>Supported values:</p>
<p>&ldquo;skip&rdquo; - Default. Nothing happens; the Workflow Execution is not started.</p>
<p>&ldquo;bufferOne&rdquo; - Starts the Workflow Execution as soon as the current one completes.
The buffer is limited to one. If another Workflow Execution is supposed to start,
but one is already in the buffer, only the one in the buffer eventually starts.</p>
<p>&ldquo;bufferAll&rdquo; - Allows an unlimited number of Workflows to buffer. They are started sequentially.</p>
<p>&ldquo;cancelOther&rdquo; - Cancels the running Workflow Execution, and then starts the new one
after the old one completes cancellation.</p>
<p>&ldquo;terminateOther&rdquo; - Terminates the running Workflow Execution and starts the new one immediately.</p>
<p>&ldquo;allowAll&rdquo; - Starts any number of concurrent Workflow Executions.
With this policy (and only this policy), more than one Workflow Execution,
started by the Schedule, can run simultaneously.</p>
<h3 id="temporal.io/v1beta1.SchedulePolicies">SchedulePolicies
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.Schedule">Schedule</a>)
</p>
<p>SchedulePolicies represent policies for overlaps, catchups, pause on failure, and workflow ID.</p>
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
<code>overlap</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleOverlapPolicy">
ScheduleOverlapPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>catchupWindow</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CatchupWindow The Temporal Server might be down or unavailable at the time
when a Schedule should take an Action. When the Server comes back up,
CatchupWindow controls which missed Actions should be taken at that point.</p>
</td>
</tr>
<tr>
<td>
<code>pauseOnFailure</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>PauseOnFailure if true, and a workflow run fails or times out, turn on &ldquo;paused&rdquo;.
This applies after retry policies: the full chain of retries must fail to trigger a pause here.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleSecondMinuteRange">ScheduleSecondMinuteRange
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">ScheduleCalendarSpec</a>)
</p>
<p>If end &lt; start, then end is interpreted as
equal to start. This means you can use a Range with start set to a value, and
end and step unset to represent a single value.</p>
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
<code>start</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Start of the range (inclusive).
Defaults to 0.</p>
</td>
</tr>
<tr>
<td>
<code>end</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>End of the range (inclusive).
Defaults to start.</p>
</td>
</tr>
<tr>
<td>
<code>step</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Step to be take between each value.
Defaults to 1.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleSpec">ScheduleSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.Schedule">Schedule</a>)
</p>
<p>ScheduleSpec is a complete description of a set of absolute timestamps.</p>
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
<code>calendars</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">
[]ScheduleCalendarSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Calendars represents calendar-based specifications of times.</p>
</td>
</tr>
<tr>
<td>
<code>intervals</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleIntervalSpec">
[]ScheduleIntervalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Intervals represents interval-based specifications of times.</p>
</td>
</tr>
<tr>
<td>
<code>crons</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Crons are cron based specifications of times.
Crons is provided for easy migration from legacy Cron Workflows. For new
use cases, we recommend using ScheduleSpec.Calendars or ScheduleSpec.
Intervals for readability and maintainability. Once a schedule is created all
expressions in Crons will be translated to ScheduleSpec.Calendars on the server.</p>
<p>For example, <code>0 12 * * MON-WED,FRI</code> is every M/Tu/W/F at noon</p>
<p>The string can have 5, 6, or 7 fields, separated by spaces, and they are interpreted in the
same way as a ScheduleCalendarSpec:</p>
<pre><code>- 5 fields:         Minute, Hour, DayOfMonth, Month, DayOfWeek
- 6 fields:         Minute, Hour, DayOfMonth, Month, DayOfWeek, Year
- 7 fields: Second, Minute, Hour, DayOfMonth, Month, DayOfWeek, Year
</code></pre>
<p>Notes:
- If Year is not given, it defaults to *.
- If Second is not given, it defaults to 0.
- Shorthands @yearly, @monthly, @weekly, @daily, and @hourly are also
accepted instead of the 5-7 time fields.
- @every <interval>[/<phase>] is accepted and gets compiled into an
IntervalSpec instead. <interval> and <phase> should be a decimal integer
with a unit suffix s, m, h, or d.
- Optionally, the string can be preceded by CRON_TZ=<time zone name> or
TZ=<time zone name>, which will get copied to ScheduleSpec.TimeZoneName. (In which case the ScheduleSpec.TimeZone field should be left empty.)
- Optionally, &ldquo;#&rdquo; followed by a comment can appear at the end of the string.
- Note that the special case that some cron implementations have for
treating DayOfMonth and DayOfWeek as &ldquo;or&rdquo; instead of &ldquo;and&rdquo; when both
are set is not implemented.</p>
</td>
</tr>
<tr>
<td>
<code>excludeCalendars</code><br>
<em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">
[]ScheduleCalendarSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExcludeCalendars defines any matching times that will be skipped.</p>
<p>All fields of the ScheduleCalendarSpec including seconds must match a time for the time to be skipped.</p>
</td>
</tr>
<tr>
<td>
<code>startAt</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartAt represents the start of the schedule. Any times before <code>startAt</code> will be skipped.
Together, <code>startAt</code> and <code>endAt</code> make an inclusive interval.
Defaults to the beginning of time.
For example: 2024-05-13T00:00:00Z</p>
</td>
</tr>
<tr>
<td>
<code>endAt</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EndAt represents the end of the schedule. Any times after <code>endAt</code> will be skipped.
Defaults to the end of time.
For example: 2024-05-13T00:00:00Z</p>
</td>
</tr>
<tr>
<td>
<code>jitter</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Jitter represents a duration that is used to apply a jitter to scheduled times.
All times will be incremented by a random value from 0 to this amount of jitter, capped
by the time until the next schedule.
Defaults to 0.</p>
</td>
</tr>
<tr>
<td>
<code>timezoneName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TimeZoneName represents the IANA time zone name, for example <code>US/Pacific</code>.</p>
<p>The definition will be loaded by Temporal Server from the environment it runs in.</p>
<p>Calendar spec matching is based on literal matching of the clock time
with no special handling of DST: if you write a calendar spec that fires
at 2:30am and specify a time zone that follows DST, that action will not
be triggered on the day that has no 2:30am. Similarly, an action that
fires at 1:30am will be triggered twice on the day that has two 1:30s.</p>
<p>Note: No actions are taken on leap-seconds (e.g. 23:59:60 UTC).
Defaults to UTC.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleState">ScheduleState
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.Schedule">Schedule</a>)
</p>
<p>ScheduleState describes the current state of a schedule.</p>
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
<code>notes</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Note is an informative human-readable message with contextual notes, e.g. the reason
a Schedule is paused. The system may overwrite this message on certain
conditions, e.g. when pause-on-failure happens.</p>
</td>
</tr>
<tr>
<td>
<code>paused</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Paused is true if the schedule is paused.</p>
</td>
</tr>
<tr>
<td>
<code>limitedActions</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>LimitedActions limits actions. While true RemainingActions will be decremented for each action taken.
Skipped actions (due to overlap policy) do not count against remaining actions.</p>
</td>
</tr>
<tr>
<td>
<code>remainingActions</code><br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>RemainingActions represents the Actions remaining in this Schedule.
Once this number hits 0, no further Actions are taken.
manual actions through backfill or ScheduleHandle.Trigger still run.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleWorkflowAction">ScheduleWorkflowAction
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleAction">ScheduleAction</a>)
</p>
<p>ScheduleWorkflowAction describes a workflow to launch.</p>
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
<code>id</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>WorkflowID represents the business identifier of the workflow execution.
The WorkflowID of the started workflow may not match this exactly,
it may have a timestamp appended for uniqueness.
Defaults to a uuid.</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br>
<em>
string
</em>
</td>
<td>
<p>WorkflowType represents the identifier used by a workflow author to define the workflow
Workflow type name.</p>
</td>
</tr>
<tr>
<td>
<code>taskQueue</code><br>
<em>
string
</em>
</td>
<td>
<p>TaskQueue represents a workflow task queue.
This is also the name of the activity task queue on which activities are scheduled.</p>
</td>
</tr>
<tr>
<td>
<code>inputs</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Inputs contains arguments to pass to the workflow.</p>
</td>
</tr>
<tr>
<td>
<code>executionTimeout</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>WorkflowExecutionTimeout is the timeout for duration of workflow execution.</p>
</td>
</tr>
<tr>
<td>
<code>runTimeout</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<p>WorkflowRunTimeout is the timeout for duration of a single workflow run.</p>
</td>
</tr>
<tr>
<td>
<code>taskTimeout</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>WorkflowTaskTimeout is The timeout for processing workflow task from the time the worker
pulled this task.</p>
</td>
</tr>
<tr>
<td>
<code>retryPolicy</code><br>
<em>
<a href="#temporal.io/v1beta1.RetryPolicy">
RetryPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RetryPolicy is the retry policy for the workflow. If a retry policy is specified,
in case of workflow failure server will start new workflow execution if
needed based on the retry policy.</p>
</td>
</tr>
<tr>
<td>
<code>memo</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Memo is optional non-indexed info that will be shown in list workflow.</p>
</td>
</tr>
<tr>
<td>
<code>searchAttributes</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SearchAttributes is optional indexed info that can be used in query of List/Scan/Count workflow APIs. The key
and value type must be registered on Temporal server side. For supported operations on different server versions
see <a href="https://docs.temporal.io/visibility">Visibility</a>.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ScheduleYearRange">ScheduleYearRange
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ScheduleCalendarSpec">ScheduleCalendarSpec</a>)
</p>
<p>If end &lt; start, then end is interpreted as
equal to start. This means you can use a Range with start set to a value, and
end and step unset to represent a single value.</p>
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
<code>start</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Start of the range (inclusive).</p>
</td>
</tr>
<tr>
<td>
<code>end</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>End of the range (inclusive).
Defaults to start.</p>
</td>
</tr>
<tr>
<td>
<code>step</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Step to be take between each value.
Defaults to 1.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.SecretKeyReference">SecretKeyReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.DatastoreSpec">DatastoreSpec</a>, 
<a href="#temporal.io/v1beta1.DatastoreTLSSpec">DatastoreTLSSpec</a>)
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
<h3 id="temporal.io/v1beta1.ServiceSpec">ServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.InternalFrontendServiceSpec">InternalFrontendServiceSpec</a>, 
<a href="#temporal.io/v1beta1.ServicesSpec">ServicesSpec</a>)
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
<p>Port defines a custom gRPC port for the service.
Default values are:
7233 for Frontend service
7234 for History service
7235 for Matching service
7239 for Worker service</p>
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
<p>MembershipPort defines a custom membership port for the service.
Default values are:
6933 for Frontend service
6934 for History service
6935 for Matching service
6939 for Worker service</p>
</td>
</tr>
<tr>
<td>
<code>httpPort</code><br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTPPort defines a custom http port for the service.
Default values are:
7243 for Frontend service</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Number of desired replicas for the service. Default to 1.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Compute Resources required by this service.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>overrides</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceSpecOverride">
ServiceSpecOverride
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides adds some overrides to the resources deployed for the service.
Those overrides takes precedence over spec.services.overrides.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>InitContainers adds a list of init containers to the service&rsquo;s deployment.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ServiceSpecOverride">ServiceSpecOverride
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.ServiceSpec">ServiceSpec</a>, 
<a href="#temporal.io/v1beta1.ServicesSpec">ServicesSpec</a>, 
<a href="#temporal.io/v1beta1.TemporalAdminToolsSpec">TemporalAdminToolsSpec</a>, 
<a href="#temporal.io/v1beta1.TemporalUISpec">TemporalUISpec</a>)
</p>
<p>ServiceSpecOverride provides the ability to override the generated manifests of a temporal service.</p>
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
<code>deployment</code><br>
<em>
<a href="#temporal.io/v1beta1.DeploymentOverride">
DeploymentOverride
</a>
</em>
</td>
<td>
<p>Override configuration for the temporal service Deployment.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ServiceStatus">ServiceStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterStatus">TemporalClusterStatus</a>)
</p>
<p>ServiceStatus reports a service status.</p>
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
<p>Name of the temporal service.</p>
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
<p>Current observed version of the service.</p>
</td>
</tr>
<tr>
<td>
<code>ready</code><br>
<em>
bool
</em>
</td>
<td>
<p>Ready defines if the service is ready.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.ServicesSpec">ServicesSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>ServicesSpec contains all temporal services specifications.</p>
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
<a href="#temporal.io/v1beta1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Frontend service custom specifications.</p>
</td>
</tr>
<tr>
<td>
<code>internalFrontend</code><br>
<em>
<a href="#temporal.io/v1beta1.InternalFrontendServiceSpec">
InternalFrontendServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Internal Frontend service custom specifications.
Only compatible with temporal &gt;= 1.20.0</p>
</td>
</tr>
<tr>
<td>
<code>history</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>History service custom specifications.</p>
</td>
</tr>
<tr>
<td>
<code>matching</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Matching service custom specifications.</p>
</td>
</tr>
<tr>
<td>
<code>worker</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Worker service custom specifications.</p>
</td>
</tr>
<tr>
<td>
<code>overrides</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceSpecOverride">
ServiceSpecOverride
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides adds some overrides to the resources deployed for all temporal services services.
Those overrides can be customized per service using spec.services.<serviceName>.overrides.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalAdminToolsSpec">TemporalAdminToolsSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>TemporalAdminToolsSpec defines parameters for the temporal admin tools within a Temporal cluster deployment.
Note that deployed admin tools version is the same as the cluster&rsquo;s version.</p>
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled defines if the operator should deploy the admin tools alongside the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image defines the temporal admin tools docker image the instance should run.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Compute Resources required by the ui.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>overrides</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceSpecOverride">
ServiceSpecOverride
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides adds some overrides to the resources deployed for the ui.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalClusterClient">TemporalClusterClient
</h3>
<p>A TemporalClusterClient creates a new mTLS client in the targeted temporal cluster.</p>
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
<a href="#temporal.io/v1beta1.TemporalClusterClientSpec">
TemporalClusterClientSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>clusterRef</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalReference">
TemporalReference
</a>
</em>
</td>
<td>
<p>Reference to the temporal cluster the client will get access to.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalClusterClientStatus">
TemporalClusterClientStatus
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
<h3 id="temporal.io/v1beta1.TemporalClusterClientSpec">TemporalClusterClientSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterClient">TemporalClusterClient</a>)
</p>
<p>TemporalClusterClientSpec defines the desired state of ClusterClient.</p>
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
<code>clusterRef</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalReference">
TemporalReference
</a>
</em>
</td>
<td>
<p>Reference to the temporal cluster the client will get access to.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalClusterClientStatus">TemporalClusterClientStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterClient">TemporalClusterClient</a>)
</p>
<p>TemporalClusterClientStatus defines the observed state of ClusterClient.</p>
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
<code>serverName</code><br>
<em>
string
</em>
</td>
<td>
<p>ServerName is the hostname returned by the certificate.</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Reference to the Kubernetes Secret containing the certificate for the client.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalCluster">TemporalCluster</a>)
</p>
<p>TemporalClusterSpec defines the desired state of Cluster.</p>
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
<p>Image defines the temporal server docker image the cluster should use for each services.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br>
<em>
github.com/alexandrevilain/temporal-operator/pkg/version.Version
</em>
</td>
<td>
<em>(Optional)</em>
<p>Version defines the temporal version the cluster to be deployed.
This version impacts the underlying persistence schemas versions.</p>
</td>
</tr>
<tr>
<td>
<code>log</code><br>
<em>
<a href="#temporal.io/v1beta1.LogSpec">
LogSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Log defines temporal cluster&rsquo;s logger configuration.</p>
</td>
</tr>
<tr>
<td>
<code>jobTtlSecondsAfterFinished</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>JobTTLSecondsAfterFinished is amount of time to keep job pods after jobs are completed.
Defaults to 300 seconds.</p>
</td>
</tr>
<tr>
<td>
<code>jobResources</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>JobResources allows set resources for setup/update jobs.</p>
</td>
</tr>
<tr>
<td>
<code>jobInitContainers</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>JobInitContainers adds a list of init containers to the setup&rsquo;s jobs.</p>
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
<a href="#temporal.io/v1beta1.ServicesSpec">
ServicesSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Services allows customizations for each temporal services deployment.</p>
</td>
</tr>
<tr>
<td>
<code>persistence</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalPersistenceSpec">
TemporalPersistenceSpec
</a>
</em>
</td>
<td>
<p>Persistence defines temporal persistence configuration.</p>
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
<p>An optional list of references to secrets in the same namespace
to use for pulling temporal images from registries.</p>
</td>
</tr>
<tr>
<td>
<code>ui</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalUISpec">
TemporalUISpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>UI allows configuration of the optional temporal web ui deployed alongside the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>admintools</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalAdminToolsSpec">
TemporalAdminToolsSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdminTools allows configuration of the optional admin tool pod deployed alongside the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>mTLS</code><br>
<em>
<a href="#temporal.io/v1beta1.MTLSSpec">
MTLSSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MTLS allows configuration of the network traffic encryption for the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>metrics</code><br>
<em>
<a href="#temporal.io/v1beta1.MetricsSpec">
MetricsSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metrics allows configuration of scraping endpoints for stats. prometheus or m3.</p>
</td>
</tr>
<tr>
<td>
<code>dynamicConfig</code><br>
<em>
<a href="#temporal.io/v1beta1.DynamicConfigSpec">
DynamicConfigSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DynamicConfig allows advanced configuration for the temporal cluster.</p>
</td>
</tr>
<tr>
<td>
<code>archival</code><br>
<em>
<a href="#temporal.io/v1beta1.ClusterArchivalSpec">
ClusterArchivalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Archival allows Workflow Execution Event Histories and Visibility data backups for the temporal cluster.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br>
<em>
<a href="#temporal.io/v1beta1.AuthorizationSpec">
AuthorizationSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization allows authorization configuration for the temporal cluster.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalClusterStatus">TemporalClusterStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalCluster">TemporalCluster</a>)
</p>
<p>TemporalClusterStatus defines the observed state of Cluster.</p>
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
<code>version</code><br>
<em>
string
</em>
</td>
<td>
<p>Version holds the current temporal version.</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceStatus">
[]ServiceStatus
</a>
</em>
</td>
<td>
<p>Services holds all services statuses.</p>
</td>
</tr>
<tr>
<td>
<code>persistence</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalPersistenceStatus">
TemporalPersistenceStatus
</a>
</em>
</td>
<td>
<p>Persistence holds all datastores statuses.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<p>Conditions represent the latest available observations of the Cluster state.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalNamespace">TemporalNamespace
</h3>
<p>A TemporalNamespace creates a namespace in the targeted temporal cluster.</p>
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
<a href="#temporal.io/v1beta1.TemporalNamespaceSpec">
TemporalNamespaceSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>clusterRef</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalReference">
TemporalReference
</a>
</em>
</td>
<td>
<p>Reference to the temporal cluster the namespace will be created.</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace description.</p>
</td>
</tr>
<tr>
<td>
<code>ownerEmail</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace owner email.</p>
</td>
</tr>
<tr>
<td>
<code>retentionPeriod</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<p>RetentionPeriod to apply on closed workflow executions.</p>
</td>
</tr>
<tr>
<td>
<code>data</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Data is a key-value map for any customized purpose.</p>
</td>
</tr>
<tr>
<td>
<code>securityToken</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>isGlobalNamespace</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>IsGlobalNamespace defines whether the namespace is a global namespace.</p>
</td>
</tr>
<tr>
<td>
<code>clusters</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of clusters names to which the namespace can fail over.
Only applicable if the namespace is a global namespace.</p>
</td>
</tr>
<tr>
<td>
<code>activeClusterName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of active Temporal Cluster.
Only applicable if the namespace is a global namespace.</p>
</td>
</tr>
<tr>
<td>
<code>allowDeletion</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowDeletion makes the controller delete the Temporal namespace if the
CRD is deleted.</p>
</td>
</tr>
<tr>
<td>
<code>archival</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalNamespaceArchivalSpec">
TemporalNamespaceArchivalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Archival is a per-namespace archival configuration.
If not set, the default cluster configuration is used.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalNamespaceStatus">
TemporalNamespaceStatus
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
<h3 id="temporal.io/v1beta1.TemporalNamespaceArchivalSpec">TemporalNamespaceArchivalSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalNamespaceSpec">TemporalNamespaceSpec</a>)
</p>
<p>TemporalNamespaceArchivalSpec is a per-namespace archival configuration override.</p>
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
<code>history</code><br>
<em>
<a href="#temporal.io/v1beta1.ArchivalSpec">
ArchivalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>History is the config for this namespace history archival.</p>
</td>
</tr>
<tr>
<td>
<code>visibility</code><br>
<em>
<a href="#temporal.io/v1beta1.ArchivalSpec">
ArchivalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Visibility is the config for this namespace visibility archival.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalNamespaceSpec">TemporalNamespaceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalNamespace">TemporalNamespace</a>)
</p>
<p>TemporalNamespaceSpec defines the desired state of Namespace.</p>
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
<code>clusterRef</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalReference">
TemporalReference
</a>
</em>
</td>
<td>
<p>Reference to the temporal cluster the namespace will be created.</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace description.</p>
</td>
</tr>
<tr>
<td>
<code>ownerEmail</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace owner email.</p>
</td>
</tr>
<tr>
<td>
<code>retentionPeriod</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<p>RetentionPeriod to apply on closed workflow executions.</p>
</td>
</tr>
<tr>
<td>
<code>data</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Data is a key-value map for any customized purpose.</p>
</td>
</tr>
<tr>
<td>
<code>securityToken</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>isGlobalNamespace</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>IsGlobalNamespace defines whether the namespace is a global namespace.</p>
</td>
</tr>
<tr>
<td>
<code>clusters</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of clusters names to which the namespace can fail over.
Only applicable if the namespace is a global namespace.</p>
</td>
</tr>
<tr>
<td>
<code>activeClusterName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of active Temporal Cluster.
Only applicable if the namespace is a global namespace.</p>
</td>
</tr>
<tr>
<td>
<code>allowDeletion</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowDeletion makes the controller delete the Temporal namespace if the
CRD is deleted.</p>
</td>
</tr>
<tr>
<td>
<code>archival</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalNamespaceArchivalSpec">
TemporalNamespaceArchivalSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Archival is a per-namespace archival configuration.
If not set, the default cluster configuration is used.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalNamespaceStatus">TemporalNamespaceStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalNamespace">TemporalNamespace</a>)
</p>
<p>TemporalNamespaceStatus defines the observed state of Namespace.</p>
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
<code>conditions</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<p>Conditions represent the latest available observations of the Namespace state.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalPersistenceSpec">TemporalPersistenceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
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
<a href="#temporal.io/v1beta1.DatastoreSpec">
DatastoreSpec
</a>
</em>
</td>
<td>
<p>DefaultStore holds the default datastore specs.</p>
</td>
</tr>
<tr>
<td>
<code>visibilityStore</code><br>
<em>
<a href="#temporal.io/v1beta1.DatastoreSpec">
DatastoreSpec
</a>
</em>
</td>
<td>
<p>VisibilityStore holds the visibility datastore specs.</p>
</td>
</tr>
<tr>
<td>
<code>secondaryVisibilityStore</code><br>
<em>
<a href="#temporal.io/v1beta1.DatastoreSpec">
DatastoreSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecondaryVisibilityStore holds the secondary visibility datastore specs.
Feature only available for clusters &gt;= 1.21.0.</p>
</td>
</tr>
<tr>
<td>
<code>advancedVisibilityStore</code><br>
<em>
<a href="#temporal.io/v1beta1.DatastoreSpec">
DatastoreSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdvancedVisibilityStore holds the advanced visibility datastore specs.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalPersistenceStatus">TemporalPersistenceStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterStatus">TemporalClusterStatus</a>)
</p>
<p>TemporalPersistenceStatus contains temporal persistence status.</p>
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
<a href="#temporal.io/v1beta1.DatastoreStatus">
DatastoreStatus
</a>
</em>
</td>
<td>
<p>DefaultStore holds the default datastore status.</p>
</td>
</tr>
<tr>
<td>
<code>visibilityStore</code><br>
<em>
<a href="#temporal.io/v1beta1.DatastoreStatus">
DatastoreStatus
</a>
</em>
</td>
<td>
<p>VisibilityStore holds the visibility datastore status.</p>
</td>
</tr>
<tr>
<td>
<code>secondaryVisibilityStore</code><br>
<em>
<a href="#temporal.io/v1beta1.DatastoreStatus">
DatastoreStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecondaryVisibilityStore holds the secondary visibility datastore status.</p>
</td>
</tr>
<tr>
<td>
<code>advancedVisibilityStore</code><br>
<em>
<a href="#temporal.io/v1beta1.DatastoreStatus">
DatastoreStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdvancedVisibilityStore holds the advanced visibility datastore status.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalReference">TemporalReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterClientSpec">TemporalClusterClientSpec</a>, 
<a href="#temporal.io/v1beta1.TemporalNamespaceSpec">TemporalNamespaceSpec</a>, 
<a href="#temporal.io/v1beta1.TemporalScheduleSpec">TemporalScheduleSpec</a>)
</p>
<p>TemporalReference is a reference to a object.</p>
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
<p>The name of the temporal object to reference.</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code><br>
<em>
string
</em>
</td>
<td>
<p>The namespace of the temporal object to reference.
Defaults to the namespace of the requested resource if omitted.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalSchedule">TemporalSchedule
</h3>
<p>A TemporalSchedule creates a schedule in the targeted temporal cluster.</p>
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
<a href="#temporal.io/v1beta1.TemporalScheduleSpec">
TemporalScheduleSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>namespaceRef</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalReference">
TemporalReference
</a>
</em>
</td>
<td>
<p>Reference to the temporal namespace the schedule will be created in.</p>
</td>
</tr>
<tr>
<td>
<code>schedule</code><br>
<em>
<a href="#temporal.io/v1beta1.Schedule">
Schedule
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>memo</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Memo is optional non-indexed info that will be shown in list workflow.</p>
</td>
</tr>
<tr>
<td>
<code>searchAttributes</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SearchAttributes is optional indexed info that can be used in query of List/Scan/Count workflow APIs. The key
and value type must be registered on Temporal server side. For supported operations on different server versions
see <a href="https://docs.temporal.io/visibility">Visibility</a>.</p>
</td>
</tr>
<tr>
<td>
<code>allowDeletion</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowDeletion makes the controller delete the Temporal schedule if the
CRD is deleted.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalScheduleStatus">
TemporalScheduleStatus
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
<h3 id="temporal.io/v1beta1.TemporalScheduleSpec">TemporalScheduleSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalSchedule">TemporalSchedule</a>)
</p>
<p>TemporalScheduleSpec defines the desired state of Schedule.</p>
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
<code>namespaceRef</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalReference">
TemporalReference
</a>
</em>
</td>
<td>
<p>Reference to the temporal namespace the schedule will be created in.</p>
</td>
</tr>
<tr>
<td>
<code>schedule</code><br>
<em>
<a href="#temporal.io/v1beta1.Schedule">
Schedule
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>memo</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Memo is optional non-indexed info that will be shown in list workflow.</p>
</td>
</tr>
<tr>
<td>
<code>searchAttributes</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SearchAttributes is optional indexed info that can be used in query of List/Scan/Count workflow APIs. The key
and value type must be registered on Temporal server side. For supported operations on different server versions
see <a href="https://docs.temporal.io/visibility">Visibility</a>.</p>
</td>
</tr>
<tr>
<td>
<code>allowDeletion</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowDeletion makes the controller delete the Temporal schedule if the
CRD is deleted.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalScheduleStatus">TemporalScheduleStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalSchedule">TemporalSchedule</a>)
</p>
<p>TemporalScheduleStatus defines the observed state of Schedule.</p>
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
<code>conditions</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<p>Conditions represent the latest available observations of the Schedule state.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalUIIngressSpec">TemporalUIIngressSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalUISpec">TemporalUISpec</a>)
</p>
<p>TemporalUIIngressSpec contains all configurations options for the UI ingress.</p>
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
<code>annotations</code><br>
<em>
map[string]string
</em>
</td>
<td>
<p>Annotations allows custom annotations on the ingress resource.</p>
</td>
</tr>
<tr>
<td>
<code>ingressClassName</code><br>
<em>
string
</em>
</td>
<td>
<p>IngressClassName is the name of the IngressClass the deployed ingress resource should use.</p>
</td>
</tr>
<tr>
<td>
<code>hosts</code><br>
<em>
[]string
</em>
</td>
<td>
<p>Host is the list of host the ingress should use.</p>
</td>
</tr>
<tr>
<td>
<code>tls</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#ingresstls-v1-networking">
[]Kubernetes networking/v1.IngressTLS
</a>
</em>
</td>
<td>
<p>TLS configuration.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="temporal.io/v1beta1.TemporalUISpec">TemporalUISpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#temporal.io/v1beta1.TemporalClusterSpec">TemporalClusterSpec</a>)
</p>
<p>TemporalUISpec defines parameters for the temporal UI within a Temporal cluster deployment.</p>
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
<code>enabled</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled defines if the operator should deploy the web ui alongside the cluster.</p>
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
<em>(Optional)</em>
<p>Version defines the temporal ui version the instance should run.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image defines the temporal ui docker image the instance should run.</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Number of desired replicas for the ui. Default to 1.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Compute Resources required by the ui.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>overrides</code><br>
<em>
<a href="#temporal.io/v1beta1.ServiceSpecOverride">
ServiceSpecOverride
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides adds some overrides to the resources deployed for the ui.</p>
</td>
</tr>
<tr>
<td>
<code>ingress</code><br>
<em>
<a href="#temporal.io/v1beta1.TemporalUIIngressSpec">
TemporalUIIngressSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ingress is an optional ingress configuration for the UI.
If lived empty, no ingress configuration will be created and the UI will only by available trough ClusterIP service.</p>
</td>
</tr>
<tr>
<td>
<code>service</code><br>
<em>
<a href="#temporal.io/v1beta1.ObjectMetaOverride">
ObjectMetaOverride
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Service is an optional service resource configuration for the UI.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
