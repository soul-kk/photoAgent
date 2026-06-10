import { Tabs } from 'expo-router';

export default function TabLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarStyle: { display: 'none' },
      }}
    >
      <Tabs.Screen name="index" />
      <Tabs.Screen name="analysis" />
      <Tabs.Screen name="advice" />
      <Tabs.Screen name="profile" />
      <Tabs.Screen name="compare" />
      <Tabs.Screen name="tone" />
    </Tabs>
  );
}
