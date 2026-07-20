import { useEffect, useState } from 'react';
import { ActivityIndicator, FlatList, SafeAreaView, StyleSheet, Text, View } from 'react-native';
import { StatusBar } from 'expo-status-bar';
import { createApiClient } from '@store/api-client';
import { formatMoney } from '@store/shared-core';

const api = createApiClient(process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1');

type Product = {
  id: string;
  name: string;
  skus?: { sale_price_cents: number }[];
};

export default function App() {
  const [items, setItems] = useState<Product[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .listProducts()
      .then((res) => setItems((res.items ?? []) as Product[]))
      .catch((e: Error) => setError(e.message));
  }, []);

  return (
    <SafeAreaView style={styles.container}>
      <StatusBar style="auto" />
      <Text style={styles.title}>Store (mobile)</Text>
      {error && <Text style={styles.error}>{error}</Text>}
      {!items.length && !error ? (
        <ActivityIndicator />
      ) : (
        <FlatList
          data={items}
          keyExtractor={(p) => p.id}
          renderItem={({ item }) => (
            <View style={styles.row}>
              <Text>{item.name}</Text>
              {item.skus?.[0] && <Text>{formatMoney(item.skus[0].sale_price_cents)}</Text>}
            </View>
          )}
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, padding: 16, gap: 8 },
  title: { fontSize: 22, fontWeight: '600', marginBottom: 8 },
  error: { color: '#b00020' },
  row: { paddingVertical: 8, borderBottomWidth: 1, borderColor: '#eee' },
});
