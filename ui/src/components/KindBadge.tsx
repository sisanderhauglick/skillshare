import Badge from './Badge';

interface KindBadgeProps {
  kind: 'skill' | 'agent';
}

export default function KindBadge({ kind }: KindBadgeProps) {
  if (kind === 'agent') {
    return <Badge variant="accent">Agent</Badge>;
  }
  return <Badge variant="info">Skill</Badge>;
}
