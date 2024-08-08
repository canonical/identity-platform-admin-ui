import axios from "axios";
import { apiBasePath } from "util/basePaths";

export const axiosInstance = axios.create({ baseURL: apiBasePath });
