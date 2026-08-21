import { z } from "zod";
export interface Floor {
    id: string;
    name: string;
}
export declare const floorSchema: z.ZodObject<{
    name: z.ZodString;
}, z.core.$strip>;
export type FloorFormData = z.infer<typeof floorSchema>;
