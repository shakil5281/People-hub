import { z } from "zod";
export interface TempShift {
    id: number;
    employee: string;
    employeeId: string;
    shift: string;
    fromDate: string;
    toDate: string;
    reason: string;
    status: "Pending" | "Approved" | "Rejected";
}
export declare const tempShiftSchema: z.ZodObject<{
    employee: z.ZodString;
    employeeId: z.ZodString;
    shift: z.ZodString;
    fromDate: z.ZodString;
    toDate: z.ZodString;
    reason: z.ZodString;
    status: z.ZodEnum<{
        Approved: "Approved";
        Pending: "Pending";
        Rejected: "Rejected";
    }>;
}, z.core.$strip>;
export type TempShiftFormData = z.infer<typeof tempShiftSchema>;
export declare function getTempShifts(): TempShift[];
export declare function getTempShift(id: number): TempShift | undefined;
export declare function createTempShift(data: TempShiftFormData): TempShift;
export declare function updateTempShift(id: number, data: Partial<TempShiftFormData>): TempShift | null;
export declare function deleteTempShift(id: number): boolean;
export declare const tempShiftStatusOptions: ({
    value: "Pending";
    label: string;
} | {
    value: "Approved";
    label: string;
} | {
    value: "Rejected";
    label: string;
})[];
export declare const shiftOptions: {
    value: string;
    label: string;
}[];
