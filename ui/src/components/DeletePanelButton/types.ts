import { ReactNode } from "react";

export type Props = {
  confirmButtonLabel: string;
  confirmButtonDisabled?: boolean;
  confirmTitle?: string;
  confirmContent: ReactNode;
  entityName: string;
  onDelete: () => Promise<unknown>;
  successMessage: string;
  successPath: string;
  invalidateQuery: string;
};

export enum Label {
  DELETE = "Delete",
}
