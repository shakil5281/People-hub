export interface Division {
    id: string;
    name: string;
    name_bn: string;
}
export interface District {
    id: string;
    name: string;
    name_bn: string;
    division_id: string;
}
export interface Upazila {
    id: string;
    name: string;
    name_bn: string;
    district_id: string;
}
export interface Union {
    id: string;
    name: string;
    name_bn: string;
    upazila_id: string;
}
