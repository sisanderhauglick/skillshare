import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Plug,
  Plus,
  RefreshCw,
  Trash2,
  X,
  EyeOff,
  Eye,
  Terminal,
  Globe,
  Server,
} from 'lucide-react';
import { api } from '../api/client';
import type { MCPServer, MCPTargetStatus, MCPCustomTarget, MCPSyncResult } from '../api/client';
import { queryKeys, staleTimes } from '../lib/queryKeys';
import { useToast } from '../components/Toast';
import Card from '../components/Card';
import Button from '../components/Button';
import Badge from '../components/Badge';
import IconButton from '../components/IconButton';
import { Input } from '../components/Input';
import DialogShell from '../components/DialogShell';
import ConfirmDialog from '../components/ConfirmDialog';
import EmptyState from '../components/EmptyState';
import PageHeader from '../components/PageHeader';
import { PageSkeleton } from '../components/Skeleton';

// ─── AddMCPServerModal ────────────────────────────────────────────────────────

interface EnvEntry {
  key: string;
  value: string;
}

function AddMCPServerModal({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: () => void;
}) {
  const { toast } = useToast();
  const [name, setName] = useState('');
  const [transport, setTransport] = useState<'stdio' | 'remote'>('stdio');
  const [command, setCommand] = useState('');
  const [args, setArgs] = useState('');
  const [url, setUrl] = useState('');
  const [envEntries, setEnvEntries] = useState<EnvEntry[]>([]);
  const [saving, setSaving] = useState(false);

  const addEnv = () => setEnvEntries((prev) => [...prev, { key: '', value: '' }]);
  const updateEnv = (i: number, field: keyof EnvEntry, value: string) => {
    setEnvEntries((prev) => prev.map((e, idx) => (idx === i ? { ...e, [field]: value } : e)));
  };
  const removeEnv = (i: number) => {
    setEnvEntries((prev) => prev.filter((_, idx) => idx !== i));
  };

  const handleCreate = async () => {
    if (!name.trim()) {
      toast('Name is required', 'error');
      return;
    }
    if (transport === 'stdio' && !command.trim()) {
      toast('Command is required for stdio transport', 'error');
      return;
    }
    if (transport === 'remote' && !url.trim()) {
      toast('URL is required for remote transport', 'error');
      return;
    }

    const validEnv = envEntries.filter((e) => e.key.trim());
    const envMap: Record<string, string> = {};
    for (const e of validEnv) {
      envMap[e.key.trim()] = e.value;
    }

    setSaving(true);
    try {
      const payload: Parameters<typeof api.createMCP>[0] = { name: name.trim() };
      if (transport === 'stdio') {
        payload.command = command.trim();
        if (args.trim()) {
          payload.args = args.trim().split(/\s+/);
        }
      } else {
        payload.url = url.trim();
      }
      if (validEnv.length > 0) {
        payload.env = envMap;
      }

      await api.createMCP(payload);
      toast(`MCP server "${name.trim()}" added`, 'success');
      onCreated();
    } catch (err: any) {
      toast(err.message ?? 'Failed to add server', 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <DialogShell open={true} onClose={onClose} maxWidth="2xl" preventClose={saving}>
      <Card overflow className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-xl font-bold text-pencil">Add MCP Server</h3>
          <IconButton
            icon={<X size={20} strokeWidth={2.5} />}
            label="Close"
            size="sm"
            variant="ghost"
            onClick={onClose}
            disabled={saving}
          />
        </div>

        <div className="space-y-4">
          {/* Name */}
          <Input
            label="Name"
            placeholder="e.g. my-mcp-server"
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={saving}
          />

          {/* Transport */}
          <div>
            <label className="block text-base text-pencil-light mb-2">Transport</label>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => setTransport('stdio')}
                disabled={saving}
                className={`flex items-center gap-2 px-4 py-2 border-2 rounded-[var(--radius-md)] text-sm font-medium transition-all ${
                  transport === 'stdio'
                    ? 'border-pencil text-pencil bg-muted/30'
                    : 'border-muted text-pencil-light hover:border-muted-dark'
                }`}
              >
                <Terminal size={14} strokeWidth={2.5} />
                stdio
              </button>
              <button
                type="button"
                onClick={() => setTransport('remote')}
                disabled={saving}
                className={`flex items-center gap-2 px-4 py-2 border-2 rounded-[var(--radius-md)] text-sm font-medium transition-all ${
                  transport === 'remote'
                    ? 'border-pencil text-pencil bg-muted/30'
                    : 'border-muted text-pencil-light hover:border-muted-dark'
                }`}
              >
                <Globe size={14} strokeWidth={2.5} />
                remote
              </button>
            </div>
          </div>

          {/* stdio fields */}
          {transport === 'stdio' && (
            <>
              <Input
                label="Command"
                placeholder="e.g. npx"
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                disabled={saving}
              />
              <Input
                label="Args (space-separated)"
                placeholder="e.g. -y @modelcontextprotocol/server-filesystem /path"
                value={args}
                onChange={(e) => setArgs(e.target.value)}
                disabled={saving}
              />
            </>
          )}

          {/* remote fields */}
          {transport === 'remote' && (
            <Input
              label="URL"
              placeholder="e.g. https://mcp.example.com/sse"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              disabled={saving}
            />
          )}

          {/* Env vars */}
          <div>
            <label className="block text-base text-pencil-light mb-2">
              Environment Variables
            </label>
            {envEntries.length > 0 && (
              <div className="space-y-2 mb-2">
                {envEntries.map((e, i) => (
                  <div key={i} className="flex gap-2 items-center">
                    <div className="flex-1">
                      <Input
                        placeholder="KEY"
                        value={e.key}
                        onChange={(ev) => updateEnv(i, 'key', ev.target.value)}
                        disabled={saving}
                      />
                    </div>
                    <div className="flex-1">
                      <Input
                        placeholder="value"
                        value={e.value}
                        onChange={(ev) => updateEnv(i, 'value', ev.target.value)}
                        disabled={saving}
                      />
                    </div>
                    <IconButton
                      icon={<X size={16} strokeWidth={2.5} />}
                      label="Remove env var"
                      size="sm"
                      variant="ghost"
                      onClick={() => removeEnv(i)}
                      disabled={saving}
                      className="hover:text-danger shrink-0"
                    />
                  </div>
                ))}
              </div>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={addEnv}
              disabled={saving}
            >
              <Plus size={14} strokeWidth={2.5} /> Add Variable
            </Button>
          </div>
        </div>

        <div className="flex gap-3 justify-end mt-6">
          <Button variant="secondary" size="sm" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button variant="primary" size="sm" onClick={handleCreate} disabled={saving}>
            {saving ? 'Adding...' : 'Add Server'}
          </Button>
        </div>
      </Card>
    </DialogShell>
  );
}

// ─── AddMCPTargetModal ────────────────────────────────────────────────────────

function AddMCPTargetModal({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: () => void;
}) {
  const { toast } = useToast();
  const [name, setName] = useState('');
  const [globalConfig, setGlobalConfig] = useState('');
  const [projectConfig, setProjectConfig] = useState('');
  const [format, setFormat] = useState<'json' | 'toml'>('json');
  const [key, setKey] = useState('mcpServers');
  const [saving, setSaving] = useState(false);

  const handleCreate = async () => {
    if (!name.trim()) {
      toast('Name is required', 'error');
      return;
    }
    if (!globalConfig.trim() && !projectConfig.trim()) {
      toast('At least one of global config path or project config path is required', 'error');
      return;
    }

    setSaving(true);
    try {
      await api.createMCPTarget({
        name: name.trim(),
        global_config: globalConfig.trim() || undefined,
        project_config: projectConfig.trim() || undefined,
        key: key.trim() || 'mcpServers',
        format,
      });
      toast(`Custom target "${name.trim()}" added`, 'success');
      onCreated();
    } catch (err: any) {
      toast(err.message ?? 'Failed to add target', 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <DialogShell open={true} onClose={onClose} maxWidth="2xl" preventClose={saving}>
      <Card overflow className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-xl font-bold text-pencil">Add Custom Target</h3>
          <IconButton
            icon={<X size={20} strokeWidth={2.5} />}
            label="Close"
            size="sm"
            variant="ghost"
            onClick={onClose}
            disabled={saving}
          />
        </div>

        <div className="space-y-4">
          <Input
            label="Name"
            placeholder="e.g. my-tool"
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={saving}
          />
          <Input
            label="Global Config Path"
            placeholder="e.g. ~/.my-tool/mcp.json"
            value={globalConfig}
            onChange={(e) => setGlobalConfig(e.target.value)}
            disabled={saving}
          />
          <Input
            label="Project Config Path"
            placeholder="e.g. .my-tool/mcp.json"
            value={projectConfig}
            onChange={(e) => setProjectConfig(e.target.value)}
            disabled={saving}
          />

          {/* Format */}
          <div>
            <label className="block text-base text-pencil-light mb-2">Format</label>
            <div className="flex gap-3">
              {(['json', 'toml'] as const).map((f) => (
                <button
                  key={f}
                  type="button"
                  onClick={() => setFormat(f)}
                  disabled={saving}
                  className={`px-4 py-2 border-2 rounded-[var(--radius-md)] text-sm font-medium transition-all ${
                    format === f
                      ? 'border-pencil text-pencil bg-muted/30'
                      : 'border-muted text-pencil-light hover:border-muted-dark'
                  }`}
                >
                  {f}
                </button>
              ))}
            </div>
          </div>

          <Input
            label="Key"
            placeholder="mcpServers"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            disabled={saving}
          />
        </div>

        <div className="flex gap-3 justify-end mt-6">
          <Button variant="secondary" size="sm" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button variant="primary" size="sm" onClick={handleCreate} disabled={saving}>
            {saving ? 'Adding...' : 'Add Target'}
          </Button>
        </div>
      </Card>
    </DialogShell>
  );
}

// ─── ServerCard ───────────────────────────────────────────────────────────────

function ServerCard({
  name,
  server,
  onToggle,
  onRemove,
}: {
  name: string;
  server: MCPServer;
  onToggle: (name: string, disabled: boolean) => Promise<void>;
  onRemove: (name: string) => void;
}) {
  const [toggling, setToggling] = useState(false);

  const handleToggle = async () => {
    setToggling(true);
    try {
      await onToggle(name, !server.disabled);
    } finally {
      setToggling(false);
    }
  };

  const isRemote = !!server.url;
  const envKeys = server.env ? Object.keys(server.env) : [];

  return (
    <Card overflow>
      {/* Header */}
      <div className="flex items-start justify-between gap-2 mb-3">
        <div className="flex items-center gap-2 flex-wrap min-w-0">
          <Plug size={15} strokeWidth={2.5} className="text-blue shrink-0" />
          <span className="font-bold text-pencil truncate">{name}</span>
          {server.disabled && <Badge variant="warning">disabled</Badge>}
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          <IconButton
            icon={
              server.disabled ? (
                <Eye size={15} strokeWidth={2.5} />
              ) : (
                <EyeOff size={15} strokeWidth={2.5} />
              )
            }
            label={server.disabled ? 'Enable server' : 'Disable server'}
            size="sm"
            variant="ghost"
            onClick={handleToggle}
            disabled={toggling}
          />
          <IconButton
            icon={<Trash2 size={15} strokeWidth={2.5} />}
            label="Remove server"
            size="sm"
            variant="danger-outline"
            onClick={() => onRemove(name)}
          />
        </div>
      </div>

      {/* Transport */}
      <div className="flex items-center gap-1.5">
        {isRemote ? (
          <Globe size={12} strokeWidth={2.5} className="text-muted-dark shrink-0" />
        ) : (
          <Terminal size={12} strokeWidth={2.5} className="text-muted-dark shrink-0" />
        )}
        <span className="text-xs text-pencil-light uppercase tracking-wider">
          {isRemote ? 'Remote' : 'stdio'}
        </span>
      </div>
      {isRemote ? (
        <p className="font-mono text-sm text-pencil-light truncate ml-4 mt-0.5">{server.url}</p>
      ) : (
        <p className="font-mono text-sm text-pencil-light truncate ml-4 mt-0.5">
          {server.command}
          {server.args && server.args.length > 0 && (
            <span className="text-muted-dark"> {server.args.join(' ')}</span>
          )}
        </p>
      )}

      {/* Env keys */}
      {envKeys.length > 0 && (
        <div className="mt-2 flex items-center gap-1.5 flex-wrap">
          <span className="text-xs text-muted-dark">env:</span>
          {envKeys.map((k) => (
            <Badge key={k} variant="default" size="sm">
              {k}
            </Badge>
          ))}
        </div>
      )}

      {/* Target filter */}
      {server.targets && server.targets.length > 0 && (
        <div className="mt-2 flex items-center gap-1.5 flex-wrap">
          <span className="text-xs text-muted-dark">targets:</span>
          {server.targets.map((t) => (
            <Badge key={t} variant="info" size="sm">
              {t}
            </Badge>
          ))}
        </div>
      )}
    </Card>
  );
}

// ─── SyncResultCard ───────────────────────────────────────────────────────────

function statusVariant(status: string): 'success' | 'danger' | 'warning' | 'default' {
  if (status === 'ok' || status === 'synced') return 'success';
  if (status === 'error' || status === 'failed') return 'danger';
  if (status === 'skipped' || status === 'dry_run') return 'warning';
  return 'default';
}

function TargetStatusCard({
  target,
  isCustom,
  onRemove,
}: {
  target: MCPTargetStatus;
  isCustom?: boolean;
  onRemove?: (name: string) => void;
}) {
  return (
    <Card>
      <div className="flex items-start justify-between gap-2 mb-2">
        <div className="flex items-center gap-2 min-w-0 flex-wrap">
          <Server size={14} strokeWidth={2.5} className="text-blue shrink-0" />
          <span className="font-bold text-pencil truncate">{target.name}</span>
          {isCustom && (
            <Badge variant="default" size="sm">
              custom
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          <Badge variant={statusVariant(target.status)} size="sm">
            {target.status}
          </Badge>
          {isCustom && onRemove && (
            <IconButton
              icon={<Trash2 size={13} strokeWidth={2.5} />}
              label="Remove custom target"
              size="sm"
              variant="danger-outline"
              onClick={() => onRemove(target.name)}
            />
          )}
        </div>
      </div>
      <p className="font-mono text-xs text-muted-dark truncate ml-5 mb-2">{target.config_path}</p>
      {target.servers && target.servers.length > 0 && (
        <div className="flex items-center gap-1.5 flex-wrap ml-5">
          {target.servers.map((s) => (
            <Badge key={s} variant="default" size="sm">
              {s}
            </Badge>
          ))}
        </div>
      )}
    </Card>
  );
}

// ─── MCPPage ──────────────────────────────────────────────────────────────────

export default function MCPPage() {
  const { toast } = useToast();
  const queryClient = useQueryClient();

  const { data, isPending, error } = useQuery({
    queryKey: queryKeys.mcp,
    queryFn: () => api.listMCP(),
    staleTime: staleTimes.mcp,
  });

  const [activeTab, setActiveTab] = useState<'servers' | 'sync'>('servers');
  const [showAdd, setShowAdd] = useState(false);
  const [showAddTarget, setShowAddTarget] = useState(false);
  const [removeName, setRemoveName] = useState<string | null>(null);
  const [removing, setRemoving] = useState(false);
  const [removeTargetName, setRemoveTargetName] = useState<string | null>(null);
  const [removingTarget, setRemovingTarget] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [syncResults, setSyncResults] = useState<MCPSyncResult[] | null>(null);

  const invalidate = () => queryClient.invalidateQueries({ queryKey: queryKeys.mcp });

  const handleToggle = async (name: string, disabled: boolean) => {
    try {
      await api.updateMCP(name, { disabled });
      toast(`Server "${name}" ${disabled ? 'disabled' : 'enabled'}`, 'success');
      invalidate();
    } catch (err: any) {
      toast(err.message ?? 'Failed to update server', 'error');
    }
  };

  const handleRemove = async () => {
    if (!removeName) return;
    setRemoving(true);
    try {
      await api.deleteMCP(removeName);
      toast(`Server "${removeName}" removed`, 'success');
      invalidate();
    } catch (err: any) {
      toast(err.message ?? 'Failed to remove server', 'error');
    } finally {
      setRemoving(false);
      setRemoveName(null);
    }
  };

  const handleRemoveTarget = async () => {
    if (!removeTargetName) return;
    setRemovingTarget(true);
    try {
      await api.deleteMCPTarget(removeTargetName);
      toast(`Custom target "${removeTargetName}" removed`, 'success');
      invalidate();
    } catch (err: any) {
      toast(err.message ?? 'Failed to remove target', 'error');
    } finally {
      setRemovingTarget(false);
      setRemoveTargetName(null);
    }
  };

  const handleSync = async (force = false) => {
    setSyncing(true);
    try {
      const res = await api.syncMCP({ force });
      setSyncResults(res.results);
      const ok = res.results.filter((r) => r.status === 'ok' || r.status === 'synced').length;
      const failed = res.results.filter((r) => r.status === 'error' || r.status === 'failed').length;
      if (failed > 0 && ok === 0) {
        toast(`Sync failed — ${failed} error${failed > 1 ? 's' : ''}`, 'error');
      } else if (failed > 0) {
        toast(`Synced ${ok} targets, ${failed} error${failed > 1 ? 's' : ''}`, 'warning');
      } else {
        toast(`Synced ${ok} target${ok !== 1 ? 's' : ''}`, 'success');
      }
      invalidate();
      setActiveTab('sync');
    } catch (err: any) {
      toast(err.message ?? 'Sync failed', 'error');
    } finally {
      setSyncing(false);
    }
  };

  if (isPending) return <PageSkeleton />;
  if (error) {
    return (
      <div className="p-6 text-danger text-sm">
        Failed to load MCP data: {(error as Error).message}
      </div>
    );
  }

  const servers = data?.servers ?? {};
  const serverEntries = Object.entries(servers);
  const targets = data?.targets ?? [];
  const mcpMode = data?.mcp_mode ?? '';
  const customTargets: Record<string, MCPCustomTarget> = data?.custom_targets ?? {};

  const tabClass = (tab: 'servers' | 'sync') =>
    `px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
      activeTab === tab
        ? 'border-pencil text-pencil'
        : 'border-transparent text-pencil-light hover:text-pencil hover:border-muted-dark'
    }`;

  return (
    <div>
      <PageHeader
        icon={<Plug size={28} strokeWidth={2.5} />}
        title="MCP Servers"
        subtitle="Manage Model Context Protocol servers synced to your AI tools"
        actions={
          <Button variant="primary" size="sm" onClick={() => setShowAdd(true)}>
            <Plus size={14} strokeWidth={2.5} /> Add Server
          </Button>
        }
      />

      {/* Tabs */}
      <div className="flex border-b border-muted mb-6">
        <button className={tabClass('servers')} onClick={() => setActiveTab('servers')}>
          Servers{serverEntries.length > 0 && ` (${serverEntries.length})`}
        </button>
        <button className={tabClass('sync')} onClick={() => setActiveTab('sync')}>
          Sync Status
        </button>
      </div>

      {/* Servers Tab */}
      {activeTab === 'servers' && (
        <>
          {serverEntries.length === 0 ? (
            <EmptyState
              icon={Plug}
              title="No MCP servers"
              description="Add a server to sync it across your AI tools."
              action={
                <Button variant="primary" size="sm" onClick={() => setShowAdd(true)}>
                  <Plus size={14} strokeWidth={2.5} /> Add Server
                </Button>
              }
            />
          ) : (
            <div className="grid gap-4 sm:grid-cols-2">
              {serverEntries.map(([name, server]) => (
                <ServerCard
                  key={name}
                  name={name}
                  server={server}
                  onToggle={handleToggle}
                  onRemove={setRemoveName}
                />
              ))}
            </div>
          )}
        </>
      )}

      {/* Sync Status Tab */}
      {activeTab === 'sync' && (
        <>
          <div className="flex items-center gap-3 mb-6 flex-wrap">
            <Button
              variant="primary"
              size="sm"
              onClick={() => handleSync(false)}
              disabled={syncing}
            >
              <RefreshCw
                size={14}
                strokeWidth={2.5}
                className={syncing ? 'animate-spin' : ''}
              />
              {syncing ? 'Syncing...' : 'Sync Now'}
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setShowAddTarget(true)}
            >
              <Plus size={14} strokeWidth={2.5} /> Add Target
            </Button>
            <div className="flex items-center gap-2">
              <label className="text-sm text-pencil-light">Mode:</label>
              <select
                value={mcpMode || 'merge'}
                onChange={async (e) => {
                  try {
                    await api.setMCPMode(e.target.value);
                    invalidate();
                  } catch (err: any) {
                    toast(err.message ?? 'Failed to update mode', 'error');
                  }
                }}
                className="text-sm border border-muted rounded-[var(--radius-sm)] px-2 py-1 bg-surface text-pencil focus:outline-none focus:border-pencil"
              >
                <option value="merge">merge</option>
                <option value="symlink">symlink</option>
                <option value="copy">copy</option>
              </select>
            </div>
          </div>

          {/* Sync results (after manual sync) */}
          {syncResults && syncResults.length > 0 && (
            <div className="mb-6">
              <h4 className="text-sm font-semibold text-pencil-light uppercase tracking-wider mb-3">
                Last Sync Results
              </h4>
              <div className="grid gap-3 sm:grid-cols-2">
                {syncResults.map((r) => (
                  <Card key={r.name} padding="sm">
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-medium text-pencil text-sm truncate">{r.name}</span>
                      <Badge variant={statusVariant(r.status)} size="sm">
                        {r.status}
                      </Badge>
                    </div>
                    {r.path && (
                      <p className="font-mono text-xs text-muted-dark truncate mt-1">{r.path}</p>
                    )}
                  </Card>
                ))}
              </div>
            </div>
          )}

          {targets.length === 0 ? (
            <EmptyState
              icon={Server}
              title="No target status"
              description="Run a sync to see target status."
            />
          ) : (
            <>
              <h4 className="text-sm font-semibold text-pencil-light uppercase tracking-wider mb-3">
                Target Status
              </h4>
              <div className="grid gap-4 sm:grid-cols-2">
                {targets.map((t) => (
                  <TargetStatusCard
                    key={t.name}
                    target={t}
                    isCustom={t.name in customTargets}
                    onRemove={setRemoveTargetName}
                  />
                ))}
              </div>
            </>
          )}
        </>
      )}

      {/* Add Server Modal */}
      {showAdd && (
        <AddMCPServerModal
          onClose={() => setShowAdd(false)}
          onCreated={() => {
            invalidate();
            setShowAdd(false);
          }}
        />
      )}

      {/* Add Target Modal */}
      {showAddTarget && (
        <AddMCPTargetModal
          onClose={() => setShowAddTarget(false)}
          onCreated={() => {
            invalidate();
            setShowAddTarget(false);
          }}
        />
      )}

      {/* Remove Server Confirm */}
      <ConfirmDialog
        open={!!removeName}
        title="Remove MCP Server"
        message={`Remove "${removeName}" from all synced targets? This cannot be undone.`}
        confirmText="Remove"
        variant="danger"
        loading={removing}
        onConfirm={handleRemove}
        onCancel={() => setRemoveName(null)}
      />

      {/* Remove Target Confirm */}
      <ConfirmDialog
        open={!!removeTargetName}
        title="Remove Custom Target"
        message={`Remove custom target "${removeTargetName}"? This will stop syncing MCP servers to this target.`}
        confirmText="Remove"
        variant="danger"
        loading={removingTarget}
        onConfirm={handleRemoveTarget}
        onCancel={() => setRemoveTargetName(null)}
      />
    </div>
  );
}
