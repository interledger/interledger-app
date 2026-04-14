interface FeatureWidgetProps {
  avatar: string        // initials (e.g. "MH")
  avatarColor?: string  // background hex
  name: string
  amount: string
  amountColor?: string  // defaults to green
  note?: string
  timestamp?: string
}

/**
 * Transaction card widget used in feature sections.
 * Matches the design reference card: avatar | name + amount | note.
 */
export function FeatureWidget({
  avatar,
  avatarColor = "#e87a7a",
  name,
  amount,
  amountColor = "var(--color-interactive-primary, #22c55e)",
  note,
  timestamp,
}: FeatureWidgetProps) {
  return (
    <div className="feature-widget">
      <div className="feature-widget__avatar" style={{ background: avatarColor }}>
        {avatar}
      </div>
      <div className="feature-widget__body">
        <div className="feature-widget__row">
          <span className="feature-widget__name">{name}</span>
          <span className="feature-widget__amount" style={{ color: amountColor }}>
            {amount}
          </span>
        </div>
        {note && <p className="feature-widget__note">{note}</p>}
        {timestamp && <p className="feature-widget__timestamp">{timestamp}</p>}
      </div>
    </div>
  )
}
