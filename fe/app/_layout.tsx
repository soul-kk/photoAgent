import { DarkTheme, ThemeProvider } from '@react-navigation/native';
import {
  PlayfairDisplay_400Regular,
  PlayfairDisplay_700Bold,
} from '@expo-google-fonts/playfair-display';
import {
  DMSans_400Regular,
  DMSans_500Medium,
  DMSans_700Bold,
} from '@expo-google-fonts/dm-sans';
import {
  DMMono_400Regular,
  DMMono_500Medium,
} from '@expo-google-fonts/dm-mono';
import { useFonts } from 'expo-font';
import { Stack, useRouter } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import React from 'react';
import 'react-native-reanimated';
import { AuthProvider, useAuth } from '@/context/auth';

export const unstable_settings = {
  anchor: '(tabs)',
};

// Inner layout reads auth state — must be a child of AuthProvider
function InnerLayout({ fontsLoaded }: { fontsLoaded: boolean }) {
  const { token, isLoaded } = useAuth();
  const router = useRouter();

  React.useEffect(() => {
    if (!isLoaded || !fontsLoaded) return;
    if (!token) router.replace('/auth');
  }, [isLoaded, fontsLoaded, token]);

  if (!fontsLoaded || !isLoaded) return null;

  return (
    <ThemeProvider value={DarkTheme}>
      <Stack>
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen name="auth" options={{ headerShown: false }} />
      </Stack>
      <StatusBar style="light" />
    </ThemeProvider>
  );
}

export default function RootLayout() {
  const [fontsLoaded] = useFonts({
    PlayfairDisplay_400Regular,
    PlayfairDisplay_700Bold,
    DMSans_400Regular,
    DMSans_500Medium,
    DMSans_700Bold,
    DMMono_400Regular,
    DMMono_500Medium,
  });

  return (
    <AuthProvider>
      <InnerLayout fontsLoaded={fontsLoaded ?? false} />
    </AuthProvider>
  );
}
