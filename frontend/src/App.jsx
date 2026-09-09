import router from "./utils/router";
import { RouterProvider } from "react-router-dom";
import { VideoFeedProvider } from "./context/VideoFeedContext";

function App() {
  return (
    <VideoFeedProvider>
      <RouterProvider router={router} />
    </VideoFeedProvider>
  );
}

export default App;
