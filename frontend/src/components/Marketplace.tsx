import { Coins, Loader, Package, ShoppingCart, TrendingUp } from 'lucide-react';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { apiService } from '../services/api';
import {
    MarketplaceStats,
    Wallet
} from '../types/api';
import { BlockchainTab } from './marketplace/BlockchainTab';
import { DeliveryRequestsTab } from './marketplace/DeliveryRequestsTab';
import { PlantListingsTab } from './marketplace/PlantListingsTab';
import { WalletTab } from './marketplace/WalletTab';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Tabs, TabsContent, TabsContents, TabsList, TabsTrigger } from './animate-ui/components/tabs';

export function Marketplace() {
    const [activeTab, setActiveTab] = useState('delivery-requests');
    const [loading, setLoading] = useState(true);
    const [stats, setStats] = useState<MarketplaceStats | null>(null);
    const [wallet, setWallet] = useState<Wallet | null>(null);

    useEffect(() => {
        loadInitialData();
    }, []);

    const loadInitialData = async () => {
        try {
            setLoading(true);
            const [statsResponse, walletResponse] = await Promise.all([
                apiService.getMarketplaceStats(),
                apiService.getWalletBalance()
            ]);
            setStats(statsResponse);
            setWallet(walletResponse.wallet);
        } catch (error) {
            console.error('Failed to load marketplace data:', error);
            toast.error('Failed to load marketplace data');
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center">
                <div className="text-center">
                    <Loader className="w-8 h-8 text-green-600 animate-spin mx-auto mb-4" />
                    <p className="text-muted-foreground">Loading marketplace...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="bg-background max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-4 sm:py-6 lg:py-8 space-y-4 sm:space-y-6">
            <div className="container mx-auto px-4 py-8">
                {/* Header */}
                <div className="mb-8">
                    <h1 className="text-4xl font-bold text-foreground mb-2">Marketplace</h1>
                    <p className="text-muted-foreground text-lg">
                        Trade plants, fulfill delivery requests, and explore the blockchain-powered economy
                    </p>
                </div>

                {/* Stats Cards */}
                {stats && (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Active Requests</CardTitle>
                                <Package className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{stats.active_delivery_requests}</div>
                                <p className="text-xs text-muted-foreground">
                                    {stats.total_delivery_requests} total requests
                                </p>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Active Listings</CardTitle>
                                <ShoppingCart className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{stats.active_listings}</div>
                                <p className="text-xs text-muted-foreground">
                                    {stats.total_listings} total listings
                                </p>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Total Volume</CardTitle>
                                <TrendingUp className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{stats.total_volume.toLocaleString()}</div>
                                <p className="text-xs text-muted-foreground">
                                    {stats.total_transactions} transactions
                                </p>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Your Balance</CardTitle>
                                <Coins className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{wallet?.balance?.toLocaleString() || 0}</div>
                                <p className="text-xs text-muted-foreground">
                                    Reputation: {wallet?.reputation || 0}
                                </p>
                            </CardContent>
                        </Card>
                    </div>
                )}

                {/* Main Content */}
                <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                    <TabsList className="mb-6 bg-muted/50 backdrop-blur-sm border border-border/50">
                        <TabsTrigger value="delivery-requests" className="text-sm font-medium">
                            <Package className="h-4 w-4 mr-2" />
                            Delivery Requests
                        </TabsTrigger>
                        <TabsTrigger value="listings" className="text-sm font-medium">
                            <ShoppingCart className="h-4 w-4 mr-2" />
                            Plant Listings
                        </TabsTrigger>
                        <TabsTrigger value="blockchain" className="text-sm font-medium">
                            <TrendingUp className="h-4 w-4 mr-2" />
                            Blockchain
                        </TabsTrigger>
                        <TabsTrigger value="wallet" className="text-sm font-medium">
                            <Coins className="h-4 w-4 mr-2" />
                            Wallet
                        </TabsTrigger>
                    </TabsList>
                    <TabsContents className="mt-6">
                        <TabsContent value="delivery-requests">
                            <DeliveryRequestsTab onRefresh={loadInitialData} />
                        </TabsContent>

                        <TabsContent value="listings">
                            <PlantListingsTab onRefresh={loadInitialData} />
                        </TabsContent>

                        <TabsContent value="blockchain">
                            <BlockchainTab />
                        </TabsContent>

                        <TabsContent value="wallet">
                            <WalletTab wallet={wallet} onRefresh={loadInitialData} />
                        </TabsContent>
                    </TabsContents>
                </Tabs>
            </div>
        </div>
    );
} 