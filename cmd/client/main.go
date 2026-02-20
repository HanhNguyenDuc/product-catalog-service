package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	productv1 "github.com/product-catalog-service/gen/product/v1"
)

func printJSON(v proto.Message) {
	marshaler := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
	b, err := marshaler.Marshal(v)
	if err != nil {
		log.Fatalf("failed to marshal proto: %v", err)
	}
	fmt.Println(string(b))
}

var (
	addr = flag.String("addr", ":50051", "gRPC server address")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [global-flags] <command> [command-flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nGlobal Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  create     Create a new product\n")
		fmt.Fprintf(os.Stderr, "  update     Update an existing product\n")
		fmt.Fprintf(os.Stderr, "  get        Get a product by ID\n")
		fmt.Fprintf(os.Stderr, "  list       List products\n")
		fmt.Fprintf(os.Stderr, "  activate   Activate a product\n")
		fmt.Fprintf(os.Stderr, "  deactivate Deactivate a product\n")
		fmt.Fprintf(os.Stderr, "  discount   Manage discounts (subcommands: apply, remove)\n")
	}
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := productv1.NewProductServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := flag.Args()[0]
	args := flag.Args()[1:]

	switch cmd {
	case "create":
		createProduct(ctx, client, args)
	case "update":
		updateProduct(ctx, client, args)
	case "get":
		getProduct(ctx, client, args)
	case "list":
		listProducts(ctx, client, args)
	case "activate":
		activateProduct(ctx, client, args)
	case "deactivate":
		deactivateProduct(ctx, client, args)
	case "discount":
		manageDiscount(ctx, client, args)
	default:
		log.Fatalf("unknown command: %s", cmd)
	}
}

func createProduct(ctx context.Context, client productv1.ProductServiceClient, args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	name := fs.String("name", "", "Product name")
	desc := fs.String("desc", "", "Product description")
	cat := fs.String("cat", "", "Product category")
	fs.Parse(args)

	if *name == "" || *cat == "" {
		log.Fatal("name and category are required")
	}

	resp, err := client.CreateProduct(ctx, &productv1.CreateProductRequest{
		Name:        *name,
		Description: *desc,
		Category:    *cat,
	})
	if err != nil {
		log.Fatalf("CreateProduct failed: %v", err)
	}
	fmt.Printf("Created Product ID: %s\n", resp.Id)
}

func updateProduct(ctx context.Context, client productv1.ProductServiceClient, args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	id := fs.String("id", "", "Product ID")
	name := fs.String("name", "", "New name (optional)")
	desc := fs.String("desc", "", "New description (optional)")
	cat := fs.String("cat", "", "New category (optional)")
	price := fs.Float64("price", -1, "New base price (optional)") // -1 indicates not set
	fs.Parse(args)

	if *id == "" {
		log.Fatal("id is required")
	}

	req := &productv1.UpdateProductRequest{Id: *id}
	if *name != "" {
		req.Name = *name
	}
	if *desc != "" {
		req.Description = *desc
	}
	if *cat != "" {
		req.Category = *cat
	}
	if *price >= 0 {

	}

	_, err := client.UpdateProduct(ctx, req)
	if err != nil {
		log.Fatalf("UpdateProduct failed: %v", err)
	}
	fmt.Println("Product updated successfully")
}

func getProduct(ctx context.Context, client productv1.ProductServiceClient, args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	id := fs.String("id", "", "Product ID")
	fs.Parse(args)

	if *id == "" {
		log.Fatal("id is required")
	}

	resp, err := client.GetProduct(ctx, &productv1.GetProductRequest{Id: *id})
	if err != nil {
		log.Fatalf("GetProduct failed: %v", err)
	}
	printJSON(resp.Product)
}

func listProducts(ctx context.Context, client productv1.ProductServiceClient, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	cat := fs.String("cat", "", "Filter by category")
	limit := fs.Int("limit", 10, "Limit results")
	offset := fs.Int("offset", 0, "Offset results")
	fs.Parse(args)

	req := &productv1.ListProductsRequest{
		Limit:  int32(*limit),
		Offset: int32(*offset),
	}
	if *cat != "" {
		req.Category = *cat
	}

	resp, err := client.ListProducts(ctx, req)
	if err != nil {
		log.Fatalf("ListProducts failed: %v", err)
	}
	printJSON(resp)
}

func activateProduct(ctx context.Context, client productv1.ProductServiceClient, args []string) {
	fs := flag.NewFlagSet("activate", flag.ExitOnError)
	id := fs.String("id", "", "Product ID")
	fs.Parse(args)

	if *id == "" {
		log.Fatal("id is required")
	}

	_, err := client.ActivateProduct(ctx, &productv1.ActivateProductRequest{Id: *id})
	if err != nil {
		log.Fatalf("ActivateProduct failed: %v", err)
	}
	fmt.Println("Product activated")
}

func deactivateProduct(ctx context.Context, client productv1.ProductServiceClient, args []string) {
	fs := flag.NewFlagSet("deactivate", flag.ExitOnError)
	id := fs.String("id", "", "Product ID")
	fs.Parse(args)

	if *id == "" {
		log.Fatal("id is required")
	}

	_, err := client.DeactivateProduct(ctx, &productv1.DeactivateProductRequest{Id: *id})
	if err != nil {
		log.Fatalf("DeactivateProduct failed: %v", err)
	}
	fmt.Println("Product deactivated")
}

func manageDiscount(ctx context.Context, client productv1.ProductServiceClient, args []string) {
	if len(args) < 1 {
		log.Fatal("subcommand required: apply or remove")
	}
	sub := args[0]
	subArgs := args[1:]

	switch sub {
	case "apply":
		fs := flag.NewFlagSet("discount apply", flag.ExitOnError)
		id := fs.String("id", "", "Product ID")
		pct := fs.String("pct", "", "Percentage (e.g. 10.5)")
		dur := fs.String("duration", "24h", "Duration (e.g. 2h, 30m)")
		fs.Parse(subArgs)

		if *id == "" || *pct == "" {
			log.Fatal("id and pct are required")
		}

		duration, err := time.ParseDuration(*dur)
		if err != nil {
			log.Fatalf("invalid duration: %v", err)
		}
		startsAt := time.Now()
		endsAt := startsAt.Add(duration)

		_, err = client.ApplyDiscount(ctx, &productv1.ApplyDiscountRequest{
			Id:         *id,
			Percentage: *pct,
			StartsAt:   timestamppb.New(startsAt),
			EndsAt:     timestamppb.New(endsAt),
		})
		if err != nil {
			log.Fatalf("ApplyDiscount failed: %v", err)
		}
		fmt.Printf("Discount applied: %s%% for %s\n", *pct, *dur)

	case "remove":
		fs := flag.NewFlagSet("discount remove", flag.ExitOnError)
		id := fs.String("id", "", "Product ID")
		fs.Parse(subArgs)

		if *id == "" {
			log.Fatal("id is required")
		}

		_, err := client.RemoveDiscount(ctx, &productv1.RemoveDiscountRequest{Id: *id})
		if err != nil {
			log.Fatalf("RemoveDiscount failed: %v", err)
		}
		fmt.Println("Discount removed")

	default:
		log.Fatalf("unknown discount subcommand: %s", sub)
	}
}
