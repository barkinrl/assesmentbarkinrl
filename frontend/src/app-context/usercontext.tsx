/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import { Version } from "../types";

export interface UserContract {
  username: string;
  isAdmin: boolean;
}

export interface DataContract {
  isDirty: boolean;
  version: Version | null;
  publish: () => void;
}

export interface ErrorMessage {
  title: string;
  message: string;
}

export interface PageContract {
  isEditable: boolean;
  inEdit: boolean;
  isDirty: boolean;
  errorMessage?: ErrorMessage;
  onSave?: () => boolean;
  onCancel?: () => void;
  onInsert?: () => void;
}

export interface AppState {
  user: UserContract;
  data: DataContract;
  page: PageContract;
  updateUser: (newUser: Partial<UserContract>) => void;
  updateData: (newData: Partial<DataContract>) => void;
  updatePage: (newPage: Partial<PageContract>) => void;
}

const publishData = () => {
  console.warn("publishData not implemented");
};

export const defaultState: AppState = {
  user: { username: "guest", isAdmin: false },
  data: { isDirty: false, version: null, publish: publishData },
  page: { isEditable: false, inEdit: false, isDirty: false },
  updateUser: (newUser?: Partial<UserContract>) => {
    console.warn("updateUser not implemented: ", newUser);
  },
  updateData: (newData?: Partial<DataContract>) => {
    console.warn("updateData not implemented: ", newData);
  },
  updatePage: (newPage?: Partial<PageContract>) => {
    console.warn("updatePage not implemented: ", newPage);
  },
};

export const AppContext = React.createContext<AppState>(defaultState);
