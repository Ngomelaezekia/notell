import createServiceClient from "../../utils/serviceClient";

const baseURL = import.meta.env.VITE_PAYMENT_SERVICE_URL || "http://localhost:8085";
const paymentClient = createServiceClient(baseURL);

export const paymentAPI = {
  packages: async () => (await paymentClient.get("/v1/packages")).data,
  subscribe: async (id, giftCode = "") => (await paymentClient.post(`/v1/packages/${id}/subscribe`, giftCode ? { giftCode } : {})).data,
  checkout: async (subscriptionId) => (await paymentClient.post(`/v1/subscriptions/${subscriptionId}/checkout`)).data,
  subscriptions: async () => (await paymentClient.get("/v1/subscriptions")).data,
  entitlement: async (resourceType, resourceId) => (await paymentClient.get(`/v1/entitlements/check/${encodeURIComponent(resourceType)}/${encodeURIComponent(resourceId)}`)).data,
  claimGift: async (code) => (await paymentClient.post("/v1/gift-tokens/claim", { code })).data,
  createAffiliateGift: async () => (await paymentClient.post("/v1/gift-tokens", { type: "USER_AFFILIATION" })).data,
  myGiftTokens: async () => (await paymentClient.get("/v1/gift-tokens/mine")).data,
};

export default paymentAPI;
