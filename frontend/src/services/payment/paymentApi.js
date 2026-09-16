import axios from "axios";

const baseURL = import.meta.env.VITE_PAYMENT_SERVICE_URL || "http://localhost:8085";
const paymentClient = axios.create({ baseURL, withCredentials: true, headers: { "Content-Type": "application/json" } });

export const paymentAPI = {
  packages: async () => (await paymentClient.get("/v1/packages")).data,
  subscribe: async (id) => (await paymentClient.post(`/v1/packages/${id}/subscribe`)).data,
  checkout: async (subscriptionId) => (await paymentClient.post(`/v1/subscriptions/${subscriptionId}/checkout`)).data,
  subscriptions: async () => (await paymentClient.get("/v1/subscriptions")).data,
  entitlement: async (resourceType, resourceId) => (await paymentClient.get(`/v1/entitlements/${resourceType}/${resourceId}`)).data,
};

export default paymentAPI;
