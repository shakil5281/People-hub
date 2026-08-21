import { z } from "zod";
export interface DailyAttendance {
    id: number;
    employee: string;
    employeeId: string;
    company: string;
    department: string;
    designation: string;
    line: string;
    group: string;
    date: string;
    checkIn: string;
    checkOut: string;
    status: "Present" | "Late" | "Absent" | "Half Day" | "Holiday" | "Leave";
    late: string;
    overTime: string;
    note: string;
}
export declare const dailyAttendanceSchema: z.ZodObject<{
    employee: z.ZodString;
    employeeId: z.ZodString;
    company: z.ZodString;
    department: z.ZodString;
    designation: z.ZodString;
    line: z.ZodString;
    group: z.ZodString;
    date: z.ZodString;
    checkIn: z.ZodString;
    checkOut: z.ZodString;
    status: z.ZodEnum<{
        Present: "Present";
        Late: "Late";
        Absent: "Absent";
        "Half Day": "Half Day";
        Holiday: "Holiday";
        Leave: "Leave";
    }>;
    note: z.ZodOptional<z.ZodString>;
}, z.core.$strip>;
export type DailyAttendanceFormData = z.infer<typeof dailyAttendanceSchema>;
export declare function getDailyAttendance(): DailyAttendance[];
export declare function getDailyAttendanceRecord(id: number): DailyAttendance | undefined;
export declare function createDailyAttendance(data: DailyAttendanceFormData): DailyAttendance;
export declare function updateDailyAttendance(id: number, data: Partial<DailyAttendanceFormData>): DailyAttendance | null;
export declare function deleteDailyAttendance(id: number): boolean;
export declare const attendanceStatusOptions: ({
    value: "Present";
    label: string;
} | {
    value: "Late";
    label: string;
} | {
    value: "Absent";
    label: string;
} | {
    value: "Half Day";
    label: string;
} | {
    value: "Holiday";
    label: string;
} | {
    value: "Leave";
    label: string;
})[];
export declare const companyOptions: {
    value: string;
    label: string;
}[];
export declare const departmentOptions: {
    value: string;
    label: string;
}[];
export declare const designationOptions: {
    value: string;
    label: string;
}[];
export declare const lineOptions: {
    value: string;
    label: string;
}[];
export declare const groupOptions: {
    value: string;
    label: string;
}[];
