import { PanelHelper } from "util/usePanelParams";

export type Props = {
  openPanel: (panelParams: PanelHelper) => void;
};

export enum Label {
  EDIT = "Edit",
}
