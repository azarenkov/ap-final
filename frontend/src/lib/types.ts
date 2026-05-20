

export type Timestamp = string | { seconds?: number | string; nanos?: number };

export interface Train {
    id: string;
    code: string;
    name: string;
    route_id: string;
    departure_time: Timestamp;
    arrival_time: Timestamp;
    total_seats: number;
    available_seats: number;
    price_cents: number;
    status: string;
}

export interface SearchResponse {
    trains?: Train[];
    total?: number;
}

export interface Route {
    id: string;
    origin: string;
    destination: string;
    distance_km: number;
    estimated_minutes: number;
}

export interface Schedule {
    train_id: string;
    stops: Array<{ station: string; arrival?: Timestamp; departure?: Timestamp }>;
}

export interface Booking {
    id: string;
    user_id: string;
    train_id: string;
    seat_count: number;
    amount_cents: number;
    status: string;
    created_at: Timestamp;
    updated_at: Timestamp;
}

export interface PaymentResult {
    booking_id: string;
    success: boolean;
    message: string;
}

export interface Ticket {
    id: string;
    booking_id: string;
    code: string;
    issued_at: Timestamp;
}

export interface Notification {
    id: string;
    user_id: string;
    kind: string;
    subject: string;
    body: string;
    read: boolean;
    created_at: Timestamp;
}
