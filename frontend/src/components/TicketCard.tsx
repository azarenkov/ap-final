import { Link } from "react-router-dom";
import type { Train } from "../lib/types";
import { formatDate, formatDuration, formatPrice, formatTime } from "../lib/format";
import t from "../styles/ticket.module.css";
import a from "../styles/app.module.css";

interface Props {
    train: Train;
    origin?: string;
    destination?: string;
}

export function TicketCard({ train, origin = "Origin", destination = "Destination" }: Props) {
    const badgeCls =
        train.status === "SCHEDULED"
            ? a.badgeOk
            : train.status === "DELAYED"
              ? a.badgeWarn
              : train.status === "CANCELLED"
                ? a.badgeDanger
                : a.badgeNeutral;

    return (
        <article className={t.ticket}>
            <div className={t.body}>
                <div className={t.codeLine}>
                    <div>
                        <div className={t.code}>{train.code}</div>
                        <div className={t.train}>{train.name}</div>
                    </div>
                    <span className={`${a.badge} ${badgeCls}`}>{train.status}</span>
                </div>

                <div className={t.route}>
                    <div className={t.station}>
                        <div className={t.stationName}>{origin}</div>
                        <div className={t.stationTime}>{formatTime(train.departure_time)}</div>
                        <div className={t.stationLabel}>departure · {formatDate(train.departure_time)}</div>
                    </div>
                    <div className={t.routeMiddle}>
                        <span className={t.duration}>{formatDuration(train.departure_time, train.arrival_time)}</span>
                        <div className={t.routeLine} />
                        <svg className={t.routeIcon} width="32" height="14" viewBox="0 0 32 14" fill="none" aria-hidden>
                            <path d="M2 7h22M19 2l5 5-5 5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                    </div>
                    <div className={`${t.station} ${t.stationRight}`}>
                        <div className={t.stationName}>{destination}</div>
                        <div className={t.stationTime}>{formatTime(train.arrival_time)}</div>
                        <div className={t.stationLabel}>arrival · {formatDate(train.arrival_time)}</div>
                    </div>
                </div>

                <div className={t.bodyFooter}>
                    <span>
                        Seats <strong>{train.available_seats}</strong> / {train.total_seats}
                    </span>
                    <span>·</span>
                    <span>Route id <strong>{train.route_id.slice(0, 8)}…</strong></span>
                </div>
            </div>

            <aside className={t.stub}>
                <div className={t.stubTop}>
                    <span className={t.stubLabel}>fare</span>
                    <span className={t.stubPrice}>
                        <em>$</em>
                        {formatPrice(train.price_cents)}
                    </span>
                    <span className={t.stubLabel}>per seat · USD</span>
                </div>
                <div className={t.stubBottom}>
                    <Link to={`/trains/${train.id}`} className={`${t.stubButton} bare`}>
                        Reserve →
                    </Link>
                    <Link to={`/trains/${train.id}`} className={`${t.stubLink} bare`}>
                        view details
                    </Link>
                </div>
            </aside>
        </article>
    );
}
