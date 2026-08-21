export interface DailySummary {
    id: number;
    date: string;
    company: string;
    department: string;
    line: string;
    present: number;
    late: number;
    absent: number;
    halfDay: number;
    leave: number;
    holiday: number;
    total: number;
}
export declare function getDailySummaries(): DailySummary[];
export declare const companyOptions: {
    value: string;
    label: string;
}[];
export declare const departmentOptions: {
    value: string;
    label: string;
}[];
export declare const lineOptions: {
    value: string;
    label: string;
}[];
